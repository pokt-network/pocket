package peerstore_provider

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/peerstore_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types PeerstoreProvider

import (
	"go.uber.org/multierr"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const ModuleName = "peerstore_provider"

// PeerstoreProvider is an interface that provides Peerstore accessors
type PeerstoreProvider interface {
	modules.Module

	GetStakedPeerstoreAtHeight(height uint64) (typesP2P.Peerstore, error)
	GetP2PConfig() *configs.P2PConfig
}

func ActorsToPeerstore(abp PeerstoreProvider, actors []*coreTypes.Actor) (pstore typesP2P.Peerstore, errs error) {
	pstore = make(typesP2P.PeerAddrMap)
	for _, a := range actors {
		networkPeer, err := ActorToPeer(abp, a)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		if err = pstore.AddPeer(networkPeer); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return pstore, errs
}

func ActorToPeer(abp PeerstoreProvider, actor *coreTypes.Actor) (typesP2P.Peer, error) {
	pubKey, err := cryptoPocket.NewPublicKey(actor.GetPublicKey())
	if err != nil {
		return nil, err
	}

	peer := &typesP2P.NetworkPeer{
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceURL: actor.GetServiceUrl(), // service url
	}

	return peer, nil
}
