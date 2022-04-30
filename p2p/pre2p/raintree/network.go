package raintree

import (
	"fmt"
	"log"
	"net"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"

	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
	}
	n.handleAddrBookUpdates()
	return n
}

func (n *rainTreeNetwork) NetworkBroadcast(data []byte) error {
	// TODO(drewsky): How should we reduce the # of envelopes here
	rainTreeMsg := &typesPre2P.RainTreeMessage{
		Data:  data,
		Level: n.maxNumLevels,
	}
	anyProto, err := anypb.New(rainTreeMsg)
	if err != nil {
		return err
	}
	pocketEvent := &types.PocketEvent{
		Topic: types.PocketTopic_P2P_BROADCAST_TOPIC,
		Data:  anyProto,
	}
	bz, err := proto.Marshal(pocketEvent)
	if err != nil {
		return err
	}

	if addr1, ok := n.getFirstTargetAddr(); ok {
		n.NetworkSend(bz, addr1)
	}
	if addr2, ok := n.getSecondTargetAddr(); ok {
		n.NetworkSend(bz, addr2)
	}
	n.NetworkSend(bz, n.addr) // Demote

	return nil
}

func (n *rainTreeNetwork) NetworkSend(data []byte, address cryptoPocket.Address) error {
	peer, ok := n.addrBookMap[address.String()]
	if !ok {
		return fmt.Errorf("Address %s not found in addrBookMap", address.String())
	}
	if !ok {
	}

	client, err := net.DialTCP("tcp4", nil, peer.ConsensusAddr)
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
