package raintree

import (
	"fmt"
	"log"
	"math/rand"
	"time"

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

	selfAddr cryptoPocket.Address

	peersManager *peersManager

	cleanupLayerEnabled    bool // debug config only
	redundancyLayerEnabled bool

	// TECHDEBT(drewsky): What should we use for de-duping messages within P2P?
	mempool map[uint64]struct{} // TODO (drewsky) replace map implementation (can grow unbounded)
}

func NewRainTreeNetwork(addr cryptoPocket.Address, addrBook typesP2P.AddrBook, cfg modules.P2PConfig) typesP2P.Network {
	pm, err := newPeersManager(addr, addrBook)
	if err != nil {
		log.Println("[ERROR] Error initializing rainTreeNetwork peersManager: ", err)
	}

	n := &rainTreeNetwork{
		selfAddr:               addr,
		peersManager:           pm,
		redundancyLayerEnabled: cfg.GetUseRainTreeRedundancyLayer(),
		cleanupLayerEnabled:    cfg.GetUseRainTreeCleanupLayer(),
		mempool:                make(map[uint64]struct{}),
	}

	return typesP2P.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastAtLevel(data, int32(n.peersManager.getNetworkView().maxNumLevels), getNonce())
}

func (n *rainTreeNetwork) networkBroadcastAtLevel(data []byte, level int32, nonce uint64) error {
	msg := &typesP2P.RainTreeMessage{
		Level: level,
		Data:  data,
		Nonce: nonce,
	}

	// Redundancy layer
	if level == 0 {
		if n.redundancyLayerEnabled {
			level, msg.Level = n.redundancyLayer()
		} else {
			if err := n.demote(msg); err != nil {
				log.Println("Error demoting self during RainTree message propagation: ", err)
			}
			return nil
		}
	}

	targets := n.getTargetsAtLevel(uint32(level))

	// Cleanup layer
	if level == -1 {
		if !n.cleanupLayerEnabled {
			return nil
		}
		targets = n.cleanupLayer()
	}

	msgBz, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if shouldSendToTarget(target) {
			log.Println("[DEBUG] "+n.selfAddr.String(), "WRITE message: with level/nonce", msg.Level, "/", msg.Nonce)
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

// Cleanup layer is just another send left / right
// This layer happens at level -1
func (n *rainTreeNetwork) cleanupLayer() []target {
	// TODO (Team) unhappy path where the left / right nodes are down
	// 			   (continue to search left and right until you have a hit)
	return n.getCleanupTargets()
}

// Redundancy layer is simply a redundant send to the original +1/3 && -1/3
func (n *rainTreeNetwork) redundancyLayer() (level int32, msgLevel int32) {
	return int32(n.peersManager.maxNumLevels), 0
}

func (n *rainTreeNetwork) demote(rainTreeMsg *typesP2P.RainTreeMessage) error {
	if rainTreeMsg.Level >= 0 {
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
	if address == nil {
		return fmt.Errorf("address %s is empty, likely not found in addrBookMap", address)
	}
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

	n.GetBus().
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
	log.Println("[DEBUG] "+n.selfAddr.String(), "READ message: with level/nonce", rainTreeMsg.Level, "/", rainTreeMsg.Nonce)
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
