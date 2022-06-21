// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	"log"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
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

func (n *network) GetBus() modules.Bus  { return nil }
func (n *network) SetBus(_ modules.Bus) {}
func (n *network) Start() error         { return nil }
func (n *network) Stop() error          { return nil }
