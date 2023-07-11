package raintree

import (
	"fmt"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/config"
	"github.com/pokt-network/pocket/p2p/protocol"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/unicast"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/pocket/telemetry"
)

var (
	_ typesP2P.Router          = &rainTreeRouter{}
	_ modules.IntegrableModule = &rainTreeRouter{}
	_ rainTreeFactory          = &rainTreeRouter{}
)

type rainTreeFactory = modules.FactoryWithConfig[typesP2P.Router, *config.RainTreeConfig]

type rainTreeRouter struct {
	base_modules.IntegrableModule
	unicast.UnicastRouter

	logger *modules.Logger
	// handler is the function to call when a message is received.
	handler typesP2P.MessageHandler
	// host represents a libp2p libp2pNetwork node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
	// selfAddr is the pocket address representing this host.
	selfAddr              cryptoPocket.Address
	peersManager          *rainTreePeersManager
	pstoreProvider        peerstore_provider.PeerstoreProvider
	currentHeightProvider modules.CurrentHeightProvider
}

func NewRainTreeRouter(bus modules.Bus, cfg *config.RainTreeConfig) (typesP2P.Router, error) {
	return new(rainTreeRouter).Create(bus, cfg)
}

func (*rainTreeRouter) Create(bus modules.Bus, cfg *config.RainTreeConfig) (typesP2P.Router, error) {
	rainTreeLogger := logger.Global.CreateLoggerForModule("rainTreeRouter")
	if err := cfg.IsValid(); err != nil {
		return nil, err
	}

	rtr := &rainTreeRouter{
		host:                  cfg.Host,
		selfAddr:              cfg.Addr,
		pstoreProvider:        cfg.PeerstoreProvider,
		currentHeightProvider: cfg.CurrentHeightProvider,
		logger:                rainTreeLogger,
		handler:               cfg.Handler,
	}
	rtr.SetBus(bus)

	height := rtr.currentHeightProvider.CurrentHeight()
	pstore, err := rtr.pstoreProvider.GetStakedPeerstoreAtHeight(height)
	if err != nil {
		return nil, fmt.Errorf("getting staked peerstore at height %d: %w", height, err)
	}
	rainTreeLogger.Info().Fields(map[string]any{
		"address":        cfg.Addr,
		"host_id":        cfg.Host.ID(),
		"protocol_id":    protocol.BackgroundProtocolID,
		"current_height": height,
		"peerstore_size": pstore.Size(),
	}).Msg("initializing raintree router")

	if err := rtr.setupDependencies(); err != nil {
		return nil, err
	}

	return typesP2P.Router(rtr), nil
}

func (rtr *rainTreeRouter) Close() error {
	return nil
}

// NetworkBroadcast implements the respective member of `typesP2P.Router`.
func (rtr *rainTreeRouter) Broadcast(data []byte) error {
	return rtr.broadcastAtLevel(data, rtr.peersManager.GetMaxNumLevels())
}

// broadcastAtLevel recursively sends to both left and right target peers
// from the starting level, demoting until level == 0.
// (see: https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p)
func (rtr *rainTreeRouter) broadcastAtLevel(data []byte, level uint32) error {
	// This is handled either by the cleanup layer or redundancy layer
	if level == 0 {
		return nil
	}
	msg := &typesP2P.RainTreeMessage{
		Level: level,
		Data:  data,
	}
	msgBz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		return err
	}

	for _, target := range rtr.getTargetsAtLevel(level) {
		if shouldSendToTarget(target) {
			if err = rtr.sendInternal(msgBz, target.address); err != nil {
				rtr.logger.Error().Err(err).Msg("sending to peer during broadcast")
			}
		}
	}

	if err = rtr.demote(msg); err != nil {
		rtr.logger.Error().Err(err).Msg("demoting self during RainTree message propagation")
	}

	return nil
}

// demote broadcasts to the decremented level's targets.
// (see: https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p)
func (rtr *rainTreeRouter) demote(rainTreeMsg *typesP2P.RainTreeMessage) error {
	if rainTreeMsg.Level > 0 {
		if err := rtr.broadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1); err != nil {
			return err
		}
	}
	return nil
}

// NetworkSend implements the respective member of `typesP2P.Router`.
func (rtr *rainTreeRouter) Send(data []byte, address cryptoPocket.Address) error {
	msg := &typesP2P.RainTreeMessage{
		Level: 0, // Direct send that does not need to be propagated
		Data:  data,
	}

	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		return err
	}

	return rtr.sendInternal(bz, address)
}

// sendInternal sends `data` to the peer at pokt `address` if not self.
func (rtr *rainTreeRouter) sendInternal(data []byte, address cryptoPocket.Address) error {
	// TODO: How should we handle this?
	if rtr.selfAddr.Equals(address) {
		rtr.logger.Debug().Str("pokt_addr", address.String()).Msg("attempted to send to self")
		return nil
	}

	peer := rtr.peersManager.GetPeerstore().GetPeer(address)
	if peer == nil {
		return fmt.Errorf("%w: with pokt address %s", typesP2P.ErrUnknownPeer, address)
	}

	// debug logging
	hostname := rtr.getHostname()
	utils.LogOutgoingMsg(rtr.logger, hostname, peer)

	if err := utils.Libp2pSendToPeer(rtr.host, protocol.RaintreeProtocolID, data, peer); err != nil {
		return err
	}

	// A bus is not available In client debug mode
	bus := rtr.GetBus()
	if bus == nil {
		return nil
	}

	bus.
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			telemetry.P2P_EVENT_METRICS_NAMESPACE,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_NAME,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL, "send",
		)

	return nil
}

// handleRainTreeMsg deserializes a RainTree message to extract the `PocketEnvelope`
// bytes contained within, continues broadcast propagation, if applicable, and
// passes them off to the application by calling the configured `rtr.handler`.
// Intended to be called in a go routine.
func (rtr *rainTreeRouter) handleRainTreeMsg(rainTreeMsgBz []byte) error {
	blockHeightInt := rtr.GetBus().GetConsensusModule().CurrentHeight()
	blockHeight := fmt.Sprintf("%d", blockHeightInt)

	rtr.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			telemetry.P2P_EVENT_METRICS_NAMESPACE,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_NAME,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_HEIGHT_LABEL, blockHeight,
		)

	var rainTreeMsg typesP2P.RainTreeMessage
	if err := proto.Unmarshal(rainTreeMsgBz, &rainTreeMsg); err != nil {
		// TECHDEBT: add telemetry
		return err
	}

	if err := rtr.validateRainTreeMsg(&rainTreeMsg); err != nil {
		// TECHDEBT: add telemetry
		return fmt.Errorf("validating raintree message: %w", err)
	}

	// Continue RainTree propagation
	if rainTreeMsg.Level > 0 {
		if err := rtr.broadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1); err != nil {
			return err
		}
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if rainTreeMsg.Data == nil {
		rtr.logger.Debug().Msg("no data in RainTree message")
		return nil
	}

	// Call configured message handler with the serialized `PocketEnvelope`.
	if err := rtr.handler(rainTreeMsg.Data); err != nil {
		return fmt.Errorf("handling raintree message: %w", err)
	}
	return nil
}

// validateRainTreeMsg ensures that the `data` contained within the RainTree message
// is a valid `PocketEnvelope` by attempting to deserialize it.
func (rtr *rainTreeRouter) validateRainTreeMsg(rainTreeMsg *typesP2P.RainTreeMessage) error {
	networkMessage := messaging.PocketEnvelope{}
	return proto.Unmarshal(rainTreeMsg.Data, &networkMessage)
}

// GetPeerstore implements the respective member of `typesP2P.Router`.
func (rtr *rainTreeRouter) GetPeerstore() typesP2P.Peerstore {
	return rtr.peersManager.GetPeerstore()
}

// AddPeer implements the respective member of `typesP2P.Router`.
func (rtr *rainTreeRouter) AddPeer(peer typesP2P.Peer) error {
	// Noop if peer with the same pokt address exists in the peerstore.
	// TECHDEBT: add method(s) to update peers.
	if p := rtr.peersManager.GetPeerstore().GetPeer(peer.GetAddress()); p != nil {
		return nil
	}

	if err := utils.AddPeerToLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	rtr.peersManager.HandleEvent(
		typesP2P.PeerManagerEvent{
			EventType: typesP2P.AddPeerEventType,
			Peer:      peer,
		},
	)
	return nil
}

// RemovePeer implements the respective member of `typesP2P.Router`.
func (rtr *rainTreeRouter) RemovePeer(peer typesP2P.Peer) error {
	rtr.peersManager.HandleEvent(
		typesP2P.PeerManagerEvent{
			EventType: typesP2P.RemovePeerEventType,
			Peer:      peer,
		},
	)
	return nil
}

// Size returns the number of peers the network is aware of and would attempt to
// broadcast to.
func (rtr *rainTreeRouter) Size() int {
	return rtr.peersManager.GetPeerstore().Size()
}

// shouldSendToTarget returns false if target is self.
func shouldSendToTarget(target target) bool {
	return !target.isSelf
}

// setupUnicastRouter configures and assigns `rtr.UnicastRouter`.
func (rtr *rainTreeRouter) setupUnicastRouter() error {
	unicastRouterCfg := config.UnicastRouterConfig{
		Logger:         rtr.logger,
		Host:           rtr.host,
		ProtocolID:     protocol.RaintreeProtocolID,
		MessageHandler: rtr.handleRainTreeMsg,
		PeerHandler:    rtr.AddPeer,
	}

	unicastRouter, err := unicast.Create(rtr.GetBus(), &unicastRouterCfg)
	if err != nil {
		return fmt.Errorf("setting up unicast router: %w", err)
	}

	rtr.UnicastRouter = *unicastRouter
	return nil
}

func (rtr *rainTreeRouter) setupDependencies() error {
	if err := rtr.setupUnicastRouter(); err != nil {
		return err
	}

	pstore, err := rtr.pstoreProvider.GetStakedPeerstoreAtHeight(rtr.currentHeightProvider.CurrentHeight())
	if err != nil {
		return fmt.Errorf("getting staked peerstore: %w", err)
	}

	if err := rtr.setupPeerManager(pstore); err != nil {
		return fmt.Errorf("setting up peer manager: %w", err)
	}

	if err := utils.PopulateLibp2pHost(rtr.host, pstore); err != nil {
		return err
	}
	return nil
}

func (rtr *rainTreeRouter) setupPeerManager(pstore typesP2P.Peerstore) (err error) {
	rtr.peersManager, err = newPeersManager(rtr.selfAddr, pstore, true)
	return err
}

func (rtr *rainTreeRouter) getHostname() string {
	return rtr.GetBus().GetRuntimeMgr().GetConfig().P2P.Hostname
}
