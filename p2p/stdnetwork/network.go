// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	types "github.com/pokt-network/pocket/p2p/types"
	"log"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

var _ types.Network = &network{}

type network struct {
	addrBook types.AddrBook
}

func NewNetwork(addrBook types.AddrBook) (n types.Network) {
	return &network{
		addrBook: addrBook,
	}
}

// TODO(olshansky): How do we avoid self-broadcasts given that `AddrBook` may contain self in the current p2p implementation?
func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.GetAddrBook() {
		if err := peer.Dialer.Write(data); err != nil {
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

		if err := peer.Dialer.Write(data); err != nil {
			log.Println("Error writing to peer during send: ", err)
			return err
		}

		break
	}

	return nil
}

func (n *network) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil // intentional passthrough
}

func (n *network) GetAddrBook() types.AddrBook {
	return n.addrBook
}

func (n *network) AddPeerToAddrBook(peer *types.NetworkPeer) error {
	n.addrBook = append(n.addrBook, peer)
	return nil
}

func (n *network) RemovePeerToAddrBook(peer *types.NetworkPeer) error {
	panic("RemovePeerToAddrBook not implemented")
}
