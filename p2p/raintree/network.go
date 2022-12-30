package raintree

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pokt-network/pocket/p2p/providers"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	telemetry "github.com/pokt-network/pocket/telemetry"
	"google.golang.org/protobuf/proto"
)

var _ typesP2P.Network = &rainTreeNetwork{}
var _ modules.IntegratableModule = &rainTreeNetwork{}

type rainTreeNetwork struct {
	bus modules.Bus

	selfAddr         cryptoPocket.Address
	addrBookProvider addrbook_provider.AddrBookProvider

	peersManager *peersManager

	// TODO (#278): What should we use for de-duping messages within P2P?
	// TODO(#388): generalize to use the shared `FIFOMempool` type (in `utility/types/mempool.go` at the time of writing) in here as well for this.
	nonceSet         map[uint64]struct{}
	nonceList        []uint64
	mempoolMaxNonces uint64
	nonceSetRWMutex  sync.RWMutex
}

func NewRainTreeNetworkWithAddrBook(addr cryptoPocket.Address, addrBook typesP2P.AddrBook, p2pCfg modules.P2PConfig) typesP2P.Network {
	pm, err := newPeersManager(addr, addrBook, true)
	if err != nil {
		log.Fatalf("[ERROR] Error initializing rainTreeNetwork peersManager: %v", err)
	}

	mempoolMaxNonces := p2pCfg.GetMaxMempoolCount()
	n := &rainTreeNetwork{
		selfAddr:         addr,
		peersManager:     pm,
		nonceSet:         make(map[uint64]struct{}),
		nonceList:        make([]uint64, 0, mempoolMaxNonces),
		mempoolMaxNonces: mempoolMaxNonces,
	}

	return typesP2P.Network(n)
}

func NewRainTreeNetwork(addr cryptoPocket.Address, bus modules.Bus, p2pCfg modules.P2PConfig, addrBookProvider providers.AddrBookProvider, currentHeightProvider providers.CurrentHeightProvider) typesP2P.Network {
	addrBook, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		log.Fatalf("[ERROR] Error getting addrBook: %v", err)
	}

	pm, err := newPeersManager(addr, addrBook, true)
	if err != nil {
		log.Fatalf("[ERROR] Error initializing rainTreeNetwork peersManager: %v", err)
	}

	n := &rainTreeNetwork{
		selfAddr:         addr,
		peersManager:     pm,
		nonceSet:         make(map[uint64]struct{}),
		nonceList:        make([]uint64, 0, p2pCfg.GetMaxMempoolCount()),
		addrBookProvider: addrBookProvider,
	}
	n.SetBus(bus)
	return typesP2P.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastAtLevel(data, n.peersManager.getNetworkView().maxNumLevels, getNonce())
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
				log.Println("Error sending to peer during broadcast: ", err)
			}
		}
	}

	if err = n.demote(msg); err != nil {
		log.Println("Error demoting self during RainTree message propagation: ", err)
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
		Nonce: getNonce(),
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

	peer, ok := n.peersManager.getNetworkView().addrBookMap[address.String()]
	if !ok {
		return fmt.Errorf("address %s not found in addrBookMap", address.String())
	}

	if err := peer.Dialer.Write(data); err != nil {
		log.Println("Error writing to peer during send: ", err)
		return err
	}

	// this is because in client mode there's no bus
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
		log.Println("Error decoding network message: ", err)
		return nil, err
	}

	// Continue RainTree propagation
	if rainTreeMsg.Level > 0 {
		if err := n.networkBroadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return nil, err
		}
	}

	n.nonceSetRWMutex.RLock()
	defer n.nonceSetRWMutex.RUnlock()
	// Avoids this node from processing a messages / transactions is has already processed at the
	// application layer. The logic above makes sure it is only propagated and returns.
	// DISCUSS(#278): Add more tests to verify this is sufficient for deduping purposes.
	if _, contains := n.nonceSet[rainTreeMsg.Nonce]; contains {
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

	n.nonceSetRWMutex.Lock()
	defer n.nonceSetRWMutex.Unlock()
	n.nonceSet[rainTreeMsg.Nonce] = struct{}{}
	n.nonceList = append(n.nonceList, rainTreeMsg.Nonce)
	if uint64(len(n.nonceList)) > n.mempoolMaxNonces {
		// removing the oldest nonce and the oldest key from the slice
		delete(n.nonceSet, n.nonceList[0])
		n.nonceList = n.nonceList[1:]
	}

	// Return the data back to the caller so it can be handled by the app specific bus
	return rainTreeMsg.Data, nil
}

func (n *rainTreeNetwork) GetAddrBook() typesP2P.AddrBook {
	return n.peersManager.getNetworkView().addrBook
}

func (n *rainTreeNetwork) AddPeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	n.peersManager.wg.Add(1)
	n.peersManager.eventCh <- addressBookEvent{addToAddressBook, peer}
	n.peersManager.wg.Wait()
	return nil
}

func (n *rainTreeNetwork) RemovePeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	n.peersManager.wg.Add(1)
	n.peersManager.eventCh <- addressBookEvent{removeFromAddressBook, peer}
	n.peersManager.wg.Wait()
	return nil
}

func (n *rainTreeNetwork) SetBus(bus modules.Bus) {
	n.bus = bus
}

func (n *rainTreeNetwork) GetBus() modules.Bus {
	// TODO: Do we need this if?
	// if n.bus == nil {
	// 	log.Printf("[WARN] PocketBus is not initialized in rainTreeNetwork")
	// 	return nil
	// }
	return n.bus
}

func getNonce() uint64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Uint64()
}

// INVESTIGATE(olshansky): This did not generate a random nonce on every call

// func getNonce() uint64 {
// 	seed, err := cryptRand.Int(cryptRand.Reader, big.NewInt(math.MaxInt64))
// 	if err != nil {
// 		panic(err)
// 	}
// 	rand.Seed(seed.Int64())
// }

func shouldSendToTarget(target target) bool {
	return !target.isSelf
}
