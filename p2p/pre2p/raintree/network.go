package raintree

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"time"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"

	"google.golang.org/protobuf/proto"
)

var _ typesPre2P.Network = &rainTreeNetwork{}

type rainTreeNetwork struct {
	selfAddr cryptoPocket.Address
	addrBook typesPre2P.AddrBook

	// TODO(olshansky): still thinking through these structures
	addrBookMap  typesPre2P.AddrBookMap
	addrList     []string
	maxNumLevels uint32

	// DISCUSS(drewsky): What should we use for de-duping messages within P2P?
	mempool types.Mempool
}

func NewRainTreeNetwork(addr cryptoPocket.Address, addrBook typesPre2P.AddrBook) typesPre2P.Network {
	n := &rainTreeNetwork{
		selfAddr: addr,
		addrBook: addrBook,

		// This subset of fields are initialized by `handleAddrBookUpdates` below
		addrBookMap:  make(typesPre2P.AddrBookMap),
		addrList:     make([]string, 0),
		maxNumLevels: 0,

		// TODO: Mempool size should be configurable
		mempool: types.NewMempool(1000000, 1000),
	}

	if err := n.handleAddrBookUpdates(); err != nil {
		// DISCUSS(drewsky): if this errors, the node could still function but not participate in
		// message propagation. Should we return an error or just log?
		log.Println("[ERROR] Error initializing rainTreeNetwork: ", err)
	}

	return typesPre2P.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastInternal(data, n.maxNumLevels, getNonce())
}

func (n *rainTreeNetwork) networkBroadcastInternal(data []byte, level uint32, nonce uint64) error {
	// NOOP: Internal broadcast at height level 0
	if level == 0 {
		return nil
	}

	msg := &typesPre2P.RainTreeMessage{
		Level: level,
		Data:  data,
		Nonce: nonce,
	}
	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if addr1, ok := n.getFirstTargetAddr(level); ok {
		if err := n.networkSendInternal(bz, addr1); err != nil {
			log.Println("Error sending to peer during broadcast: ", err)
		}
	}

	if addr2, ok := n.getSecondTargetAddr(level); ok {
		if err := n.networkSendInternal(bz, addr2); err != nil {
			log.Println("Error sending to peer during broadcast: ", err)
		}
	}

	if err := n.demote(msg); err != nil {
		log.Println("Error demoting self during RainTree message propagation: ", err)
	}

	return nil
}

func (n *rainTreeNetwork) demote(rainTreeMsg *typesPre2P.RainTreeMessage) error {
	if rainTreeMsg.Level > 0 {
		if err := n.networkBroadcastInternal(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return err
		}
	}
	return nil
}

func (n *rainTreeNetwork) NetworkSend(data []byte, address cryptoPocket.Address) error {
	msg := &typesPre2P.RainTreeMessage{
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
	if n.selfAddr.Equals(address) {
		// NOOP: Trying to send a message to self
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
	var rainTreeMsg typesPre2P.RainTreeMessage
	if err := proto.Unmarshal(data, &rainTreeMsg); err != nil {
		return nil, err
	}

	networkMessage := types.PocketEvent{}
	if err := proto.Unmarshal(rainTreeMsg.Data, &networkMessage); err != nil {
		log.Println("Error decoding network message: ", err)
		return nil, err
	}

	if rainTreeMsg.Level > 0 {
		if err := n.networkBroadcastInternal(rainTreeMsg.Data, rainTreeMsg.Level-1, rainTreeMsg.Nonce); err != nil {
			return nil, err
		}
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(rainTreeMsg.Nonce))
	hash := cryptoPocket.SHA3Hash(b)
	hashString := hex.EncodeToString(hash)
	// Don't process the transaction again - only propagate
	if n.mempool.Contains(hashString) {
		return nil, nil
	}

	// Error handling the transaction
	if err := n.mempool.AddTransaction(b); err != nil {
		return nil, fmt.Errorf("error adding transaction to RainTree mempool: %s", err.Error())
	}

	// Return the data back to the caller so it can be handeled by the app specific bus
	return rainTreeMsg.Data, nil
}

func (n *rainTreeNetwork) GetAddrBook() typesPre2P.AddrBook {
	return n.addrBook

}

func (n *rainTreeNetwork) AddPeerToAddrBook(peer *typesPre2P.NetworkPeer) error {
	n.addrBook = append(n.addrBook, peer)
	if err := n.handleAddrBookUpdates(); err != nil {
		return nil
	}
	return nil
}

func (n *rainTreeNetwork) RemovePeerToAddrBook(peer *typesPre2P.NetworkPeer) error {
	panic("Not implemented")
}

func getNonce() uint64 {
	// INVESTIGATE(olshansky): This did not generate a random nonce on every call

	// seed, err := cryptRand.Int(cryptRand.Reader, big.NewInt(math.MaxInt64))
	// if err != nil {
	// 	panic(err)
	// }
	// rand.Seed(seed.Int64())

	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Uint64()
}
