// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	"fmt"

	"github.com/pokt-network/pocket/p2p/providers"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ typesP2P.Network           = &network{}
	_ modules.IntegratableModule = &network{}
)

type network struct {
	addrBookMap typesP2P.AddrBookMap

	logger modules.Logger
}

func NewNetwork(bus modules.Bus, addrBookProvider providers.AddrBookProvider, currentHeightProvider providers.CurrentHeightProvider) (n typesP2P.Network) {
	addrBook, err := addrBookProvider.GetStakedAddrBookAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		logger.Fatal().Err(err).Msg("Error getting addrBook")
	}

	addrBookMap := make(typesP2P.AddrBookMap)
	for _, peer := range addrBook {
		addrBookMap[peer.Address.String()] = peer
	}
	return &network{
		logger:      logger,
		addrBookMap: addrBookMap,
	}
}

// TODO(olshansky): How do we avoid self-broadcasts given that `AddrBook` may contain self in the current p2p implementation?
func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.addrBookMap {
		if err := peer.Dialer.Write(data); err != nil {
			n.logger.Error().Err(err).Msg("Error writing to one of the peers during broadcast")
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
		n.logger.Error().Err(err).Msg("Error writing to peer during send")
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
