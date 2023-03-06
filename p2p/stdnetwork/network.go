// TECHDEBT(olshansky): Delete this once we are fully comfortable with RainTree moving forward.

package stdnetwork

import (
	"fmt"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"

	"github.com/pokt-network/pocket/logger"
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
	pstore sharedP2P.Peerstore

	logger *modules.Logger
}

func NewNetwork(bus modules.Bus, pstoreProvider providers.PeerstoreProvider, currentHeightProvider providers.CurrentHeightProvider) (n typesP2P.Network) {
	networkLogger := logger.Global.CreateLoggerForModule("network")
	networkLogger.Info().Msg("Initializing stdnetwork")

	pstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		networkLogger.Fatal().Err(err).Msg("Error getting peerstore")
	}

	return &network{
		logger: networkLogger,
		pstore: pstore,
	}
}

// TODO(olshansky): How do we avoid self-broadcasts given that `AddrBook` may contain self in the current p2p implementation?
func (n *network) NetworkBroadcast(data []byte) error {
	for _, peer := range n.pstore.GetAllPeers() {
		if _, err := peer.GetStream().Write(data); err != nil {
			n.logger.Error().Err(err).Msg("Error writing to one of the peers during broadcast")
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

	if _, err := peer.GetStream().Write(data); err != nil {
		n.logger.Error().Err(err).Msg("Error writing to peer during send")
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
