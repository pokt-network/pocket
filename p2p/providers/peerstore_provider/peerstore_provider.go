package peerstore_provider

//go:generate mockgen -destination=../../types/mocks/peerstore_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/providers/peerstore_provider PeerstoreProvider

import (
	"go.uber.org/multierr"

	"github.com/pokt-network/pocket/logger"
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
		// TECHDEBT(#519): consider checking for behaviour instead of type. For reference: https://github.com/pokt-network/pocket/pull/611#discussion_r1147476057
		if _, ok := err.(*ErrResolvingAddr); ok {
			logger.Global.Warn().Err(err).Msg("ignoring ErrResolvingAddr - peer unreachable, not adding it to peerstore")
			continue
		} else if err != nil {
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
