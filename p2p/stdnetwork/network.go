// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	"fmt"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p/providers"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_ typesP2P.Network           = &network{}
	_ modules.IntegratableModule = &network{}
)

type network struct {
	host   libp2pHost.Host
	pstore typesP2P.Peerstore

	logger *modules.Logger
}

func NewNetwork(host libp2pHost.Host, pstoreProvider providers.PeerstoreProvider, currentHeightProvider providers.CurrentHeightProvider) (typesP2P.Network, error) {
	networkLogger := logger.Global.CreateLoggerForModule("network")
	networkLogger.Info().Msg("Initializing stdnetwork")

	pstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		return nil, err
	}

	return &network{
		host:   host,
		logger: networkLogger,
		pstore: pstore,
	}, nil
}

func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.pstore.GetPeerList() {
		if err := utils.Libp2pSendToPeer(n.host, data, peer); err != nil {
			n.logger.Error().
				Err(err).
				Bool("TODO", true).
				Str("pokt address", peer.GetAddress().String()).
				Msg("broadcasting to peer")
			continue
		}
	}
	return nil
}

func (n *network) NetworkSend(data []byte, address cryptoPocket.Address) error {
	peer := n.pstore.GetPeer(address)
	if peer == nil {
		return fmt.Errorf("peer with address %s not in peerstore", address)
	}

	if err := utils.Libp2pSendToPeer(n.host, data, peer); err != nil {
		return err
	}
	return nil
}

func (n *network) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil // intentional passthrough
}

func (n *network) GetPeerstore() typesP2P.Peerstore {
	return n.pstore
}

func (n *network) AddPeer(peer typesP2P.Peer) error {
	// Noop if peer with the pokt address already exists in the peerstore.
	// TECHDEBT: add method(s) to update peers.
	if p := n.pstore.GetPeer(peer.GetAddress()); p != nil {
		return nil
	}

	if err := utils.AddPeerToLibp2pHost(n.host, peer); err != nil {
		return err
	}

	return n.pstore.AddPeer(peer)
}

func (n *network) RemovePeer(peer typesP2P.Peer) error {
	if err := utils.RemovePeerFromLibp2pHost(n.host, peer); err != nil {
		return err
	}

	return n.pstore.RemovePeer(peer.GetAddress())
}

func (n *network) GetBus() modules.Bus  { return nil }
func (n *network) SetBus(_ modules.Bus) {}
