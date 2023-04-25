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
	_ typesP2P.Network           = &router{}
	_ modules.IntegratableModule = &router{}
)

type router struct {
	host   libp2pHost.Host
	pstore typesP2P.Peerstore

	logger *modules.Logger
}

func NewNetwork(host libp2pHost.Host, pstoreProvider providers.PeerstoreProvider, currentHeightProvider providers.CurrentHeightProvider) (typesP2P.Network, error) {
	networkLogger := logger.Global.CreateLoggerForModule("router")
	networkLogger.Info().Msg("Initializing stdnetwork")

	pstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(currentHeightProvider.CurrentHeight())
	if err != nil {
		return nil, err
	}

	return &router{
		host:   host,
		logger: networkLogger,
		pstore: pstore,
	}, nil
}

func (rtr *router) NetworkBroadcast(data []byte) error {
	for _, peer := range rtr.pstore.GetPeerList() {
		if err := utils.Libp2pSendToPeer(rtr.host, data, peer); err != nil {
			rtr.logger.Error().
				Err(err).
				Bool("TODO", true).
				Str("pokt address", peer.GetAddress().String()).
				Msg("broadcasting to peer")
			continue
		}
	}
	return nil
}

func (rtr *router) NetworkSend(data []byte, address cryptoPocket.Address) error {
	peer := rtr.pstore.GetPeer(address)
	if peer == nil {
		return fmt.Errorf("peer with address %s not in peerstore", address)
	}

	if err := utils.Libp2pSendToPeer(rtr.host, data, peer); err != nil {
		return err
	}
	return nil
}

func (rtr *router) HandleNetworkData(data []byte) ([]byte, error) {
	return data, nil // intentional passthrough
}

func (rtr *router) GetPeerstore() typesP2P.Peerstore {
	return rtr.pstore
}

func (rtr *router) AddPeer(peer typesP2P.Peer) error {
	// Noop if peer with the pokt address already exists in the peerstore.
	// TECHDEBT: add method(s) to update peers.
	if p := rtr.pstore.GetPeer(peer.GetAddress()); p != nil {
		return nil
	}

	if err := utils.AddPeerToLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	return rtr.pstore.AddPeer(peer)
}

func (rtr *router) RemovePeer(peer typesP2P.Peer) error {
	if err := utils.RemovePeerFromLibp2pHost(rtr.host, peer); err != nil {
		return err
	}

	return rtr.pstore.RemovePeer(peer.GetAddress())
}

func (rtr *router) GetBus() modules.Bus  { return nil }
func (rtr *router) SetBus(_ modules.Bus) {}
