package raintree

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
	telemetry "github.com/pokt-network/pocket/telemetry"
	"google.golang.org/protobuf/proto"
)

var _ typesP2P.Network = &rainTreeNetwork{}

type rainTreeNetwork struct {
	base_modules.IntegratableModule

	selfAddr         cryptoPocket.Address
	addrBookProvider addrbook_provider.AddrBookProvider

	peersManager *rainTreePeersManager
	nonceDeduper *mempool.GenericFIFOSet[uint64, uint64]

	currentHeightProvider providers.CurrentHeightProvider

	logger *modules.Logger
}

func NewRainTreeNetwork(addr cryptoPocket.Address, bus modules.Bus, addrBookProvider providers.AddrBookProvider, currentHeightProvider providers.CurrentHeightProvider) typesP2P.Network {
	networkLogger := logger.Global.CreateLoggerForModule("network")
	networkLogger.Info().Msg("Initializing rainTreeNetwork")

	pstore, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		networkLogger.Fatal().Err(err).Msg("Error getting pstore")
	}

	pm, err := newPeersManager(addr, pstore, true)
	if err != nil {
		networkLogger.Fatal().Err(err).Msg("Error initializing rainTreeNetwork rainTreePeersManager")
	}

	p2pCfg := bus.GetRuntimeMgr().GetConfig().P2P

	n := &rainTreeNetwork{
		selfAddr:              addr,
		peersManager:          pm,
		nonceDeduper:          mempool.NewGenericFIFOSet[uint64, uint64](int(p2pCfg.MaxMempoolCount)),
		addrBookProvider:      addrBookProvider,
		currentHeightProvider: currentHeightProvider,
		logger:                networkLogger,
	}
	n.SetBus(bus)
	return typesP2P.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastAtLevel(data, n.peersManager.GetMaxNumLevels(), crypto.GetNonce())
}

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
				n.logger.Error().Err(err).Msg("Error sending to peer during broadcast")
			}
		}
	}

	if err = n.demote(msg); err != nil {
		n.logger.Error().Err(err).Msg("Error demoting self during RainTree message propagation")
	}

	return nil
}

func (n *rainTreeNetwork) demote(rainTreeMsg *typesP2P.RainTreeMessage) error {
	if rainTreeMsg.Level > 0 {
		if err := n.networkBroadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return err
		}
	}
	return nil
}

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

func (n *rainTreeNetwork) networkSendInternal(data []byte, address cryptoPocket.Address) error {
	// NOOP: Trying to send a message to self
	if n.selfAddr.Equals(address) {
		return nil
	}

	peer := n.peersManager.GetPeersView().GetPeerstore().GetPeer(address)
	if peer == nil {
		n.logger.Error().Str("address", address.String()).Msg("address not found in peerstore")
	}

	// TECHDEBT: should not bee `Peer`s responsibility
	// to manage or expose its connections.
	if _, err := peer.GetStream().Write(data); err != nil {
		n.logger.Error().Err(err).Msg("Error writing to peer during send")
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

func (n *rainTreeNetwork) GetPeerList() sharedP2P.PeerList {
	return n.peersManager.GetPeersView().GetPeers()
}

func (n *rainTreeNetwork) AddPeer(peer sharedP2P.Peer) error {
	n.peersManager.HandleEvent(
		sharedP2P.PeerManagerEvent{
			EventType: sharedP2P.AddPeerEventType,
			Peer:      peer,
		},
	)
	return nil
}

func (n *rainTreeNetwork) RemovePeer(peer sharedP2P.Peer) error {
	n.peersManager.HandleEvent(
		sharedP2P.PeerManagerEvent{
			EventType: sharedP2P.RemovePeerEventType,
			Peer:      peer,
		},
	)
	return nil
}

func (n *rainTreeNetwork) Size() int {
	return n.peersManager.GetPeersView().GetPeerstore().Size()
}

func shouldSendToTarget(target target) bool {
	return !target.isSelf
}
