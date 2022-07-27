package raintree

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	types2 "github.com/pokt-network/pocket/p2p/types"
	"log"
	"math/rand"
	"time"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"

	"google.golang.org/protobuf/proto"
)

var _ types2.Network = &rainTreeNetwork{}

type rainTreeNetwork struct {
	selfAddr cryptoPocket.Address
	addrBook types2.AddrBook

	// TECHDEBT(olshansky): Consider optimizing these away if possible.
	// Helpers / abstractions around `addrBook` for simpler implementation through additional
	// storage & pre-computation.
	addrBookMap            types2.AddrBookMap
	addrList               []string
	maxNumLevels           int32
	redundancyLayerEnabled bool // debug config only

	// TECHDEBT(drewsky): What should we use for de-duping messages within P2P?
	mempool types.Mempool
}

func NewRainTreeNetwork(addr cryptoPocket.Address, addrBook types2.AddrBook) types2.Network {
	n := &rainTreeNetwork{
		selfAddr: addr,
		addrBook: addrBook,
		// This subset of fields are initialized by `processAddrBookUpdates` below
		addrBookMap:            make(types2.AddrBookMap),
		addrList:               make([]string, 0),
		maxNumLevels:           0,
		redundancyLayerEnabled: true,
		// TODO(team): Mempool size should be configurable
		mempool: types.NewMempool(1000000, 1000),
	}

	if err := n.processAddrBookUpdates(); err != nil {
		// DISCUSS(drewsky): if this errors, the node could still function but not participate in
		// message propagation. Should we return an error or just log?
		log.Println("[ERROR] Error initializing rainTreeNetwork: ", err)
	}

	return types2.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastAtLevel(data, n.maxNumLevels, getNonce())
}

func (n *rainTreeNetwork) networkBroadcastAtLevel(data []byte, level int32, nonce uint64) error {
	var addr1, addr2 cryptoPocket.Address
	var ok bool

	msg := &types2.RainTreeMessage{
		Level: level,
		Data:  data,
		Nonce: nonce,
	}
	// This is handled either by the redundancy layer
	if level == 0 {
		if n.redundancyLayerEnabled {
			// redundancy layer is simply one final send to the original +1/3 && -1/3
			level = n.maxNumLevels
			// ensure not an echo-chamber
			msg.Level = -1
		} else {
			if err := n.demote(msg); err != nil {
				log.Println("Error demoting self during RainTree message propagation: ", err)
			}
		}
	}

	// This is handled by the cleanup layer
	if level == -1 {
		// cleanup layer is just send left / right
		// TODO (Team) unhappy path where the left / right nodes are down
		// (continue to search left and right until you have a hit)
		addr1, addr2, ok = n.getLeftAndRight()
		if !ok {
			return nil
		}
	}

	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if addr1 == nil {
		addr1 = n.getFirstTargetAddr(level)
	}
	if addr2 == nil {
		addr2 = n.getSecondTargetAddr(level)
	}

	if err = n.networkSendInternal(bz, addr1); err != nil {
		log.Println("Error sending to peer during broadcast: ", err)
	}
	if err = n.networkSendInternal(bz, addr2); err != nil {
		log.Println("Error sending to peer during broadcast: ", err)
	}

	if err = n.demote(msg); err != nil {
		log.Println("Error demoting self during RainTree message propagation: ", err)
	}

	return nil
}

func (n *rainTreeNetwork) demote(rainTreeMsg *types2.RainTreeMessage) error {
	if rainTreeMsg.Level >= 0 {
		if err := n.networkBroadcastAtLevel(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return err
		}
	}
	return nil
}

func (n *rainTreeNetwork) NetworkSend(data []byte, address cryptoPocket.Address) error {
	msg := &types2.RainTreeMessage{
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
	if address == nil {
		return fmt.Errorf("address %s is empty, likely not found in addrBookMap", address)
	}
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
	var rainTreeMsg types2.RainTreeMessage
	if err := proto.Unmarshal(data, &rainTreeMsg); err != nil {
		return nil, err
	}

	networkMessage := types.PocketEvent{}
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

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(rainTreeMsg.Nonce))
	hash := cryptoPocket.SHA3Hash(b)
	hashString := hex.EncodeToString(hash)
	// Avoids this node from processing a messages / transactions is has already processed at the
	// application layer. The logic above makes sure it is only propagated and returns.
	// TODO(team): Add more tests to verify this is sufficient for deduping purposes.
	if n.mempool.Contains(hashString) {
		return nil, nil
	}

	// Error handling the addition transaction to the local mempool
	if err := n.mempool.AddTransaction(b); err != nil {
		return nil, fmt.Errorf("error adding transaction to RainTree mempool: %s", err.Error())
	}

	// Return the data back to the caller so it can be handeled by the app specific bus
	return rainTreeMsg.Data, nil
}

func (n *rainTreeNetwork) GetAddrBook() types2.AddrBook {
	return n.addrBook

}

func (n *rainTreeNetwork) AddPeerToAddrBook(peer *types2.NetworkPeer) error {
	n.addrBook = append(n.addrBook, peer)
	if err := n.processAddrBookUpdates(); err != nil {
		return nil
	}
	return nil
}

func (n *rainTreeNetwork) RemovePeerToAddrBook(peer *types2.NetworkPeer) error {
	panic("Not implemented")
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
