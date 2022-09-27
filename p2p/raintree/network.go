package raintree

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
	telemetry "github.com/pokt-network/pocket/telemetry"
	"google.golang.org/protobuf/proto"
)

var _ typesP2P.Network = &rainTreeNetwork{}
var _ modules.IntegratableModule = &rainTreeNetwork{}

type rainTreeNetwork struct {
	bus modules.Bus

	// TODO(olshansky): still thinking through these structures
	selfAddr cryptoPocket.Address
	addrBook typesP2P.AddrBook

	// TECHDEBT(olshansky): Consider optimizing these away if possible.
	// Helpers / abstractions around `addrBook` for simpler implementation through additional
	// storage & pre-computation.
	addrBookMap  typesP2P.AddrBookMap
	addrList     []string
	maxNumLevels uint32

	// TECHDEBT(drewsky): What should we use for de-duping messages within P2P?
	mempool map[uint64]struct{} // TODO (drewsky) replace map implementation (can grow unbounded)
}

func NewRainTreeNetwork(addr cryptoPocket.Address, addrBook typesP2P.AddrBook) typesP2P.Network {
	n := &rainTreeNetwork{
		selfAddr: addr,
		addrBook: addrBook,
		// This subset of fields are initialized by `processAddrBookUpdates` below
		addrBookMap:  make(typesP2P.AddrBookMap),
		addrList:     make([]string, 0),
		maxNumLevels: 0,
		mempool:      make(map[uint64]struct{}),
	}

	if err := n.processAddrBookUpdates(); err != nil {
		// DISCUSS(drewsky): if this errors, the node could still function but not participate in
		// message propagation. Should we return an error or just log?
		log.Println("[ERROR] Error initializing rainTreeNetwork: ", err)
	}

	return typesP2P.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastAtLevel(data, n.maxNumLevels, getNonce())
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
	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if addr1, ok := n.getFirstTargetAddr(level); ok {
		if err = n.networkSendInternal(bz, addr1); err != nil {
			log.Println("Error sending to peer during broadcast: ", err)
		}
	}

	if addr2, ok := n.getSecondTargetAddr(level); ok {
		if err = n.networkSendInternal(bz, addr2); err != nil {
			log.Println("Error sending to peer during broadcast: ", err)
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

	bz, err := proto.Marshal(msg)
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

	peer, ok := n.addrBookMap[address.String()]
	if !ok {
		return fmt.Errorf("address %s not found in addrBookMap", address.String())
	}

	if err := peer.Dialer.Write(data); err != nil {
		log.Println("Error writing to peer during send: ", err)
		return err
	}

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

	networkMessage := debug.PocketEvent{}
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

	// Avoids this node from processing a messages / transactions is has already processed at the
	// application layer. The logic above makes sure it is only propagated and returns.
	// TODO(team): Add more tests to verify this is sufficient for deduping purposes.
	if _, contains := n.mempool[rainTreeMsg.Nonce]; contains {
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

	n.mempool[rainTreeMsg.Nonce] = struct{}{}

	// Return the data back to the caller so it can be handeled by the app specific bus
	return rainTreeMsg.Data, nil
}

func (n *rainTreeNetwork) GetAddrBook() typesP2P.AddrBook {
	return n.addrBook
}

func (n *rainTreeNetwork) AddPeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	n.addrBook = append(n.addrBook, peer)
	if err := n.processAddrBookUpdates(); err != nil {
		return nil
	}
	return nil
}

func (n *rainTreeNetwork) RemovePeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	panic("Not implemented")
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
