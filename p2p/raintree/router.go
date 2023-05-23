package raintree

import (
	"fmt"
	"io"
	"log"
	"time"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/config"
	"github.com/pokt-network/pocket/p2p/protocol"
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

// TECHDEBT(#629): configure timeouts. Consider security exposure vs. real-world conditions.
// TECHDEBT(#629): parameterize and expose via config.
// readStreamTimeout is the duration to wait for a read operation on a
// stream to complete, after which the stream is closed ("timed out").
const readStreamTimeout = time.Second * 10

var (
	_ typesP2P.Router            = &rainTreeRouter{}
	_ modules.IntegratableModule = &rainTreeRouter{}
	_ rainTreeFactory            = &rainTreeRouter{}
)

type rainTreeFactory = modules.FactoryWithConfig[typesP2P.Router, *config.RainTreeConfig]

type rainTreeRouter struct {
	base_modules.IntegratableModule

	logger *modules.Logger
	// handler is the function to call when a message is received.
	handler typesP2P.RouterHandler
	// host represents a libp2p libp2pNetwork node, it encapsulates a libp2p peerstore
	// & connection manager. `libp2p.New` configures and starts listening
	// according to options.
	// (see: https://pkg.go.dev/github.com/libp2p/go-libp2p#section-readme)
	host libp2pHost.Host
	// selfAddr is the pocket address representing this host.
	selfAddr              cryptoPocket.Address
	peersManager          *rainTreePeersManager
	pstoreProvider        peerstore_provider.PeerstoreProvider
	currentHeightProvider providers.CurrentHeightProvider
	nonceDeduper          *mempool.GenericFIFOSet[uint64, uint64]
}

func NewRainTreeRouter(bus modules.Bus, cfg *config.RainTreeConfig) (typesP2P.Router, error) {
	return new(rainTreeRouter).Create(bus, cfg)
}

func (*rainTreeRouter) Create(bus modules.Bus, cfg *config.RainTreeConfig) (typesP2P.Router, error) {
	routerLogger := logger.Global.CreateLoggerForModule("router")
	routerLogger.Info().Msg("Initializing rainTreeRouter")

	if err := cfg.IsValid(); err != nil {
		return nil, err
	}

	rtr := &rainTreeRouter{
		host:                  cfg.Host,
		selfAddr:              cfg.Addr,
		nonceDeduper:          mempool.NewGenericFIFOSet[uint64, uint64](int(cfg.MaxNonces)),
		pstoreProvider:        cfg.PeerstoreProvider,
		currentHeightProvider: cfg.CurrentHeightProvider,
		logger:                routerLogger,
		handler:               cfg.Handler,
	}
	rtr.SetBus(bus)

	if err := rtr.setupDependencies(); err != nil {
		return nil, err
	}

	rtr.host.SetStreamHandler(protocol.PoktProtocolID, rtr.handleStream)
	return typesP2P.Router(rtr), nil
}

// NetworkBroadcast implements the respective member of `typesP2P.Router`.
func (rtr *rainTreeRouter) Broadcast(data []byte) error {
	return rtr.broadcastAtLevel(data, rtr.peersManager.GetMaxNumLevels(), crypto.GetNonce())
}

// broadcastAtLevel recursively sends to both left and right target peers
// from the starting level, demoting until level == 0.
// (see: https://github.com/pokt-network/pocket-network-protocol/tree/main/p2p)
func (rtr *rainTreeRouter) broadcastAtLevel(data []byte, level uint32, nonce uint64) error {
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
		if err := rtr.broadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
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
		Nonce: crypto.GetNonce(),
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
		return fmt.Errorf("no known peer with pokt address %s", address)
	}

	// debug logging
	hostname := rtr.getHostname()
	utils.LogOutgoingMsg(rtr.logger, hostname, peer)

	if err := utils.Libp2pSendToPeer(rtr.host, data, peer); err != nil {
		rtr.logger.Debug().Err(err).Msg("from libp2pSendInternal")
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

// handleRainTreeMsg handles a RainTree message, continuing broadcast propagation
// if applicable. Returns the serialized `PocketEnvelope` data contained within.
func (rtr *rainTreeRouter) handleRainTreeMsg(data []byte) ([]byte, error) {
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
	if err := proto.Unmarshal(data, &rainTreeMsg); err != nil {
		return nil, err
	}

	networkMessage := messaging.PocketEnvelope{}
	if err := proto.Unmarshal(rainTreeMsg.Data, &networkMessage); err != nil {
		rtr.logger.Error().Err(err).Msg("Error decoding network message")
		return nil, err
	}

	// Continue RainTree propagation
	if rainTreeMsg.Level > 0 {
		if err := rtr.broadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return nil, err
		}
	}

	// Avoids this node from processing a messages / transactions is has already processed at the
	// application layer. The logic above makes sure it is only propagated and returns.
	// DISCUSS(#278): Add more tests to verify this is sufficient for deduping purposes.
	if contains := rtr.nonceDeduper.Contains(rainTreeMsg.Nonce); contains {
		log.Printf("RainTree message with nonce %d already processed, skipping\n", rainTreeMsg.Nonce)
		rtr.GetBus().
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
	if err := rtr.nonceDeduper.Push(rainTreeMsg.Nonce); err != nil {
		return nil, err
	}

	// Return the data back to the caller so it can be handled by the app specific bus
	return rainTreeMsg.Data, nil
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

// handleStream ensures the peerstore contains the remote peer and then reads
// the incoming stream in a new go routine.
func (rtr *rainTreeRouter) handleStream(stream libp2pNetwork.Stream) {
	rtr.logger.Debug().Msg("handling incoming stream")
	peer, err := utils.PeerFromLibp2pStream(stream)
	if err != nil {
		rtr.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("parsing remote peer identity")

		if err = stream.Reset(); err != nil {
			rtr.logger.Error().Err(err).Msg("resetting stream")
		}
		return
	}

	if err := rtr.AddPeer(peer); err != nil {
		rtr.logger.Error().Err(err).
			Str("address", peer.GetAddress().String()).
			Msg("adding remote peer to router")
	}

	go rtr.readStream(stream)
}

// readStream reads the incoming stream, extracts the serialized `PocketEnvelope`
// data from the incoming `RainTreeMessage`, and passes it to the application by
// calling the configured `rtr.handler`.
func (rtr *rainTreeRouter) readStream(stream libp2pNetwork.Stream) {
	// Time out if no data is sent to free resources.
	if err := stream.SetReadDeadline(newReadStreamDeadline()); err != nil {
		// NB: tests using libp2p's `mocknet` rely on this not returning an error.
		// `SetReadDeadline` not supported by `mocknet` streams.
		rtr.logger.Error().Err(err).Msg("setting stream read deadline")
	}

	// log incoming stream
	rtr.logStream(stream)

	// read stream
	rainTreeMsgBz, err := io.ReadAll(stream)
	if err != nil {
		rtr.logger.Error().Err(err).Msg("reading from stream")
		if err := stream.Reset(); err != nil {
			rtr.logger.Error().Err(err).Msg("resetting stream (read-side)")
		}
		return
	}

	// done reading; reset to signal this to remote peer
	if err := stream.Reset(); err != nil {
		rtr.logger.Error().Err(err).Msg("resetting stream (read-side)")
	}

	// extract `PocketEnvelope` from `RainTreeMessage` (& continue propagation)
	poktEnvelopeBz, err := rtr.handleRainTreeMsg(rainTreeMsgBz)
	if err != nil {
		rtr.logger.Error().Err(err).Msg("handling raintree message")
		return
	}

	// There was no error, but we don't need to forward this to the app-specific bus.
	// For example, the message has already been handled by the application.
	if poktEnvelopeBz == nil {
		return
	}

	// call configured handler to forward to app-specific bus
	if err := rtr.handler(poktEnvelopeBz); err != nil {
		rtr.logger.Error().Err(err).Msg("handling pocket envelope")
	}
}

// shouldSendToTarget returns false if target is self.
func shouldSendToTarget(target target) bool {
	return !target.isSelf
}

func (rtr *rainTreeRouter) setupDependencies() error {
	pstore, err := rtr.pstoreProvider.GetStakedPeerstoreAtHeight(rtr.currentHeightProvider.CurrentHeight())
	if err != nil {
		return err
	}

	if err := rtr.setupPeerManager(pstore); err != nil {
		return err
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

// newReadStreamDeadline returns a future deadline
// based on the read stream timeout duration.
func newReadStreamDeadline() time.Time {
	return time.Now().Add(readStreamTimeout)
}
