package pre2p

import (
	"log"
	"net"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

const (
	NetworkProtocol = "tcp4"
)

var _ typesPre2P.Network = &network{}

type network struct {
	addrBook typesPre2P.AddrBook
}

func NewNetwork(addrBook typesPre2P.AddrBook) (n typesPre2P.Network) {
	return &network{
		addrBook: addrBook,
	}
}

// TODO(olshansky): How do we avoid self-broadcasts given that `AddrBook` may contain self in the current pre2p implementation?
func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.GetAddrBook() {
		client, err := net.DialTCP(NetworkProtocol, nil, peer.ConsensusAddr)
		if err != nil {
			log.Println("Error connecting to one of the peers during broadcast: ", err)
			continue
		}
		defer client.Close()

		_, err = client.Write(data)
		if err != nil {
			log.Println("Error writing to one of the peers during broadcast: ", err)
			continue
		}
	}
	return nil
}

func (n *network) NetworkSend(data []byte, address cryptoPocket.Address) error {
	for _, peer := range n.GetAddrBook() {
		// TODO(team): If the address book is a map instead of a list, we wouldn't have to do this loop.
		if address.String() != peer.PublicKey.Address().String() {
			continue
		}

		client, err := net.DialTCP(NetworkProtocol, nil, peer.ConsensusAddr)
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

		break // During a send, only one peer needs to receive the message
	}

	return nil
}

func (n *network) NetworkPropagate() typesPre2P.AddrBook {
	panic("NetworkPropagate not implemented")
}

func (n *network) GetAddrBook() typesPre2P.AddrBook {
	return n.addrBook
}

func (n *network) AddPeerToAddrBook(peer *typesPre2P.NetworkPeer) error {
	n.addrBook = append(n.addrBook, peer)
	return nil
}

func (n *network) RemovePeerToAddrBook(peer *typesPre2P.NetworkPeer) error {
	panic("RemovePeerToAddrBook not implemented")
}
