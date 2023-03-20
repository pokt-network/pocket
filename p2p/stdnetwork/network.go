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
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
)

var (
	_ typesP2P.Network           = &network{}
	_ modules.IntegratableModule = &network{}
)

type network struct {
	host   libp2pHost.Host
	pstore sharedP2P.Peerstore

	logger *modules.Logger
}

func NewNetwork(host libp2pHost.Host, pstoreProvider providers.PeerstoreProvider, currentHeightProvider providers.CurrentHeightProvider) (n typesP2P.Network, err error) {
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

func (n *network) NetworkSend(data []byte, address cryptoPocket.Address) (err error) {
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

func (n *network) GetPeerstore() sharedP2P.Peerstore {
	return n.pstore
}

func (n *network) AddPeer(peer sharedP2P.Peer) error {
	return n.pstore.AddPeer(peer)
}

func (n *network) RemovePeer(peer sharedP2P.Peer) error {
	return n.pstore.RemovePeer(peer.GetAddress())
}

func (n *network) GetBus() modules.Bus  { return nil }
func (n *network) SetBus(_ modules.Bus) {}
