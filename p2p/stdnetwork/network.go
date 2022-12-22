// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/p2p/providers"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ typesP2P.Network = &network{}
var _ modules.IntegratableModule = &network{}

type network struct {
	addrBookMap typesP2P.AddrBookMap
}

func NewNetwork(bus modules.Bus, p2pCfg modules.P2PConfig, addrBookProvider providers.AddrBookProvider, currentHeightProvider providers.CurrentHeightProvider) (n typesP2P.Network) {
	addrBook, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		log.Fatalf("[ERROR] Error getting addrBook: %v", err)
	}

	addrBookMap := make(typesP2P.AddrBookMap)
	for _, peer := range addrBook {
		addrBookMap[peer.Address.String()] = peer
	}
	return &network{
		addrBookMap: addrBookMap,
	}
}

// TODO(olshansky): How do we avoid self-broadcasts given that `AddrBook` may contain self in the current p2p implementation?
func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.addrBookMap {
		if err := peer.Dialer.Write(data); err != nil {
			log.Println("Error writing to one of the peers during broadcast: ", err)
			continue
		}
	}
	return nil
}

func (n *network) NetworkSend(data []byte, address cryptoPocket.Address) error {
	peer, ok := n.addrBookMap[address.String()]
	if !ok {
		return fmt.Errorf("peer with address %v not in addrBookMap", peer)
	}

	if err := peer.Dialer.Write(data); err != nil {
		log.Println("Error writing to peer during send: ", err)
		return err
	}

	return nil
}

func (n *network) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil // intentional passthrough
}

func (n *network) GetAddrBook() typesP2P.AddrBook {
	addrBook := make(typesP2P.AddrBook, 0)
	for _, p := range n.addrBookMap {
		addrBook = append(addrBook, p)
	}
	return addrBook
}

func (n *network) AddPeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	n.addrBookMap[peer.Address.String()] = peer
	return nil
}

func (n *network) RemovePeerToAddrBook(peer *typesP2P.NetworkPeer) error {
	delete(n.addrBookMap, peer.Address.String())
	return nil
}

func (n *network) GetBus() modules.Bus  { return nil }
func (n *network) SetBus(_ modules.Bus) {}
