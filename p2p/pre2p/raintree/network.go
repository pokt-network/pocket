package raintree

import (
	"fmt"
	"log"
	"net"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"

	"google.golang.org/protobuf/proto"
)

var _ typesPre2P.Network = &rainTreeNetwork{}

type rainTreeNetwork struct {
	addr     cryptoPocket.Address
	addrBook typesPre2P.AddrBook

	// TODO(olshansky): still thinking through these structures
	addrBookMap  typesPre2P.AddrBookMap
	addrList     []string
	maxNumLevels uint32
}

func NewRainTreeNetwork(selfAddr cryptoPocket.Address, addrBook typesPre2P.AddrBook) typesPre2P.Network {
	n := &rainTreeNetwork{
		addr:     selfAddr,
		addrBook: addrBook,

		// These fields are initialized by calling `handleAddrBookUpdates` below.
		addrBookMap:  make(typesPre2P.AddrBookMap),
		addrList:     make([]string, 0),
		maxNumLevels: 0,
	}
	n.handleAddrBookUpdates()

	return typesPre2P.Network(n)
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	return n.networkBroadcastInternal(data, n.maxNumLevels)
}

func (n *rainTreeNetwork) networkBroadcastInternal(data []byte, level uint32) error {
	msg := &typesPre2P.RainTreeMessage{
		Level: level,
		Data:  data,
	}
	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if addr1, ok := n.getFirstTargetAddr(level); ok {
		n.networkSendInternal(bz, addr1)
	}
	if addr2, ok := n.getSecondTargetAddr(level); ok {
		n.networkSendInternal(bz, addr2)
	}
	n.networkSendInternal(bz, n.addr) // Demote

	return nil
}

func (n *rainTreeNetwork) NetworkSend(data []byte, address cryptoPocket.Address) error {
	msg := &typesPre2P.RainTreeMessage{
		Level: 0, // Direct send that does not need to be propagated
		Data:  data,
	}

	bz, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return n.networkSendInternal(bz, address)
}

func (n *rainTreeNetwork) networkSendInternal(data []byte, address cryptoPocket.Address) error {
	peer, ok := n.addrBookMap[address.String()]
	if !ok {
		return fmt.Errorf("address %s not found in addrBookMap", address.String())
	}

	client, err := net.DialTCP(typesPre2P.TransportLayerProtocol, nil, peer.ConsensusAddr)
	if err != nil {
		log.Println("Error connecting to peer during send: ", err)
		return err
	}
	defer client.Close()

	_, err = client.Write(data)
	if err != nil {
		log.Println("Error writing to peer during send: ", err)
		return err
	}

	return nil
}

func (n *rainTreeNetwork) HandleRawData(data []byte) ([]byte, error) {
	var rainTreeMsg typesPre2P.RainTreeMessage
	if err := proto.Unmarshal(data, &rainTreeMsg); err != nil {
		return nil, err
	}

	// Message propagation if level is non-zero
	if rainTreeMsg.Level > 0 {
		n.networkBroadcastInternal(rainTreeMsg.Data, rainTreeMsg.Level-1)
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
