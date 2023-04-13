package raintree

import (
	"fmt"
	"log"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	telemetry "github.com/pokt-network/pocket/telemetry"
)

var (
	_ typesP2P.Network           = &rainTreeNetwork{}
	_ modules.IntegratableModule = &rainTreeNetwork{}
)

type RainTreeConfig struct {
	Addr                  cryptoPocket.Address
	PeerstoreProvider     providers.PeerstoreProvider
	CurrentHeightProvider providers.CurrentHeightProvider
	Host                  libp2pHost.Host
}

type rainTreeNetwork struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	// host represents a libp2p network node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
	// selfAddr is the pocket address representing this host.
	selfAddr cryptoPocket.Address
	// hostname is the network hostname from the config
	hostname              string
	peersManager          *rainTreePeersManager
	pstoreProvider        peerstore_provider.PeerstoreProvider
	currentHeightProvider providers.CurrentHeightProvider
	nonceDeduper          *mempool.GenericFIFOSet[uint64, uint64]
}

func NewRainTreeNetwork(bus modules.Bus, cfg RainTreeConfig) (typesP2P.Network, error) {
	return new(rainTreeNetwork).Create(bus, cfg)
}

func (*rainTreeNetwork) Create(bus modules.Bus, netCfg RainTreeConfig) (typesP2P.Network, error) {
	networkLogger := logger.Global.CreateLoggerForModule("network")
	networkLogger.Info().Msg("Initializing rainTreeNetwork")

	if err := netCfg.isValid(); err != nil {
		return nil, err
	}

	p2pCfg := bus.GetRuntimeMgr().GetConfig().P2P

	n := &rainTreeNetwork{
		host:                  netCfg.Host,
		selfAddr:              netCfg.Addr,
		hostname:              p2pCfg.Hostname,
		nonceDeduper:          mempool.NewGenericFIFOSet[uint64, uint64](int(p2pCfg.MaxMempoolCount)),
		pstoreProvider:        netCfg.PeerstoreProvider,
		currentHeightProvider: netCfg.CurrentHeightProvider,
		logger:                networkLogger,
	}
	n.SetBus(bus)

	if err := n.setupDependencies(); err != nil {
		return nil, err
	}

	return typesP2P.Network(n), nil
}

// NetworkBroadcast implements the respective member of `typesP2P.Network`.
func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastAtLevel(data, n.peersManager.GetMaxNumLevels(), crypto.GetNonce())
}

// networkBroadcastAtLevel recursively sends to both left and right target peers
// from the starting level, demoting until level == 0.
// (see: https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p)
func (n *rainTreeNetwork) networkBroadcastAtLevel(data []byte, level uint32, nonce uint64) error {
	// This is handled either by the cleanup layer or redundancy layer
	if level == 0 {
		return nil
	}
	msg := &typesP2P.RainTreeMessage{
		Level: level,
		Data:  data,
		Nonce: nonce,
	}
	msgBz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		return err
	}

	for _, target := range n.getTargetsAtLevel(level) {
		if shouldSendToTarget(target) {
			if err = n.networkSendInternal(msgBz, target.address); err != nil {
				n.logger.Error().Err(err).Msg("sending to peer during broadcast")
			}
		}
	}

	if err = n.demote(msg); err != nil {
		n.logger.Error().Err(err).Msg("demoting self during RainTree message propagation")
	}

	return nil
}

// demote broadcasts to the decremented level's targets.
// (see: https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p)
func (n *rainTreeNetwork) demote(rainTreeMsg *typesP2P.RainTreeMessage) error {
	if rainTreeMsg.Level > 0 {
		if err := n.networkBroadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return err
		}
	}
	return nil
}

// NetworkSend implements the respective member of `typesP2P.Network`.
func (n *rainTreeNetwork) NetworkSend(data []byte, address cryptoPocket.Address) error {
	msg := &typesP2P.RainTreeMessage{
		Level: 0, // Direct send that does not need to be propagated
		Data:  data,
		Nonce: crypto.GetNonce(),
	}

	bz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		return err
	}

	return n.networkSendInternal(bz, address)
}

// networkSendInternal sends `data` to the peer at pokt `address` if not self.
func (n *rainTreeNetwork) networkSendInternal(data []byte, address cryptoPocket.Address) error {
	// TODO: How should we handle this?
	if n.selfAddr.Equals(address) {
		n.logger.Debug().Str("pokt_addr", address.String()).Msg("attempted to send to self")
		return nil
	}

	peer := n.peersManager.GetPeerstore().GetPeer(address)
	if peer == nil {
		return fmt.Errorf("no known peer with pokt address %s", address)
	}

	// debug logging
	utils.LogOutgoingMsg(n.logger, n.hostname, peer)

	if err := utils.Libp2pSendToPeer(n.host, data, peer); err != nil {
		n.logger.Debug().Err(err).Msg("from libp2pSendInternal")
		return err
	}

	// A bus is not available In client debug mode
	bus := n.GetBus()
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

// HandleNetworkData implements the respective member of `typesP2P.Network`.
func (n *rainTreeNetwork) HandleNetworkData(data []byte) ([]byte, error) {
	blockHeightInt := n.GetBus().GetConsensusModule().CurrentHeight()
	blockHeight := fmt.Sprintf("%d", blockHeightInt)

	n.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			telemetry.P2P_EVENT_METRICS_NAMESPACE,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_NAME,
			telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_HEIGHT_LABEL, blockHeight,
		)

	var rainTreeMsg typesP2P.RainTreeMessage
	if err := proto.Unmarshal(data, &rainTreeMsg); err != nil {
		return nil, err
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(rainTreeMsg.Data, &networkMessage); err != nil {
		n.logger.Error().Err(err).Msg("Error decoding network message")
		return nil, err
	}

	// Continue RainTree propagation
	if rainTreeMsg.Level > 0 {
		if err := n.networkBroadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return nil, err
		}
	}

	// Avoids this node from processing a messages / transactions is has already processed at the
	// application layer. The logic above makes sure it is only propagated and returns.
	// DISCUSS(#278): Add more tests to verify this is sufficient for deduping purposes.
	if contains := n.nonceDeduper.Contains(rainTreeMsg.Nonce); contains {
		log.Printf("RainTree message with nonce %d already processed, skipping\n", rainTreeMsg.Nonce)
		n.GetBus().
			GetTelemetryModule().
			GetEventMetricsAgent().
			EmitEvent(
				telemetry.P2P_EVENT_METRICS_NAMESPACE,
				telemetry.P2P_BROADCAST_MESSAGE_REDUNDANCY_PER_BLOCK_EVENT_METRIC_NAME,
				telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_NONCE_LABEL, rainTreeMsg.Nonce,
				telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_HEIGHT_LABEL, blockHeight,
			)

		return nil, nil
	}

	// Add the nonce to the deduper
	if err := n.nonceDeduper.Push(rainTreeMsg.Nonce); err != nil {
		return nil, err
	}

	// Return the data back to the caller so it can be handled by the app specific bus
	return rainTreeMsg.Data, nil
}

// GetPeerstore implements the respective member of `typesP2P.Network`.
func (n *rainTreeNetwork) GetPeerstore() typesP2P.Peerstore {
	return n.peersManager.GetPeerstore()
}

// AddPeer implements the respective member of `typesP2P.Network`.
func (n *rainTreeNetwork) AddPeer(peer typesP2P.Peer) error {
	// Noop if peer with the same pokt address exists in the peerstore.
	// TECHDEBT: add method(s) to update peers.
	if p := n.peersManager.GetPeerstore().GetPeer(peer.GetAddress()); p != nil {
		return nil
	}

	if err := utils.AddPeerToLibp2pHost(n.host, peer); err != nil {
		return err
	}

	n.peersManager.HandleEvent(
		typesP2P.PeerManagerEvent{
			EventType: typesP2P.AddPeerEventType,
			Peer:      peer,
		},
	)
	return nil
}

func (n *rainTreeNetwork) RemovePeer(peer typesP2P.Peer) error {
	n.peersManager.HandleEvent(
		typesP2P.PeerManagerEvent{
			EventType: typesP2P.RemovePeerEventType,
			Peer:      peer,
		},
	)
	return nil
}

// Size returns the number of peers the network is aware of and would attempt to
// broadcast to.
func (n *rainTreeNetwork) Size() int {
	return n.peersManager.GetPeerstore().Size()
}

// shouldSendToTarget returns false if target is self.
func shouldSendToTarget(target target) bool {
	return !target.isSelf
}

func (n *rainTreeNetwork) setupDependencies() error {
	pstore, err := n.pstoreProvider.GetStakedPeerstoreAtHeight(n.currentHeightProvider.CurrentHeight())
	if err != nil {
		return err
	}

	if err := n.setupPeerManager(pstore); err != nil {
		return err
	}

	if err := utils.PopulateLibp2pHost(n.host, pstore); err != nil {
		return err
	}
	return nil
}

func (n *rainTreeNetwork) setupPeerManager(pstore typesP2P.Peerstore) (err error) {
	n.peersManager, err = newPeersManager(n.selfAddr, pstore, true)
	return err
}

func (cfg RainTreeConfig) isValid() (err error) {
	if cfg.Host == nil {
		err = multierr.Append(err, fmt.Errorf("host not configured"))
	}

	if cfg.PeerstoreProvider == nil {
		err = multierr.Append(err, fmt.Errorf("peerstore provider not configured"))
	}

	if cfg.CurrentHeightProvider == nil {
		err = multierr.Append(err, fmt.Errorf("current height provider not configured"))
	}
	return err
}
