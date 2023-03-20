package peerstore_provider

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/peerstore_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types PeerstoreProvider

import (
	"fmt"

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
	GetConnFactory() typesP2P.ConnectionFactory
	GetP2PConfig() *configs.P2PConfig
	SetConnectionFactory(typesP2P.ConnectionFactory)
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
	// TECHDEBT: this should be the responsibility of some new `ConnManager` interface.
	// Peerstore (identity / address mapping) is a separate concern from managing
	// connections to/from peers.
	conn, err := abp.GetConnFactory()(abp.GetP2PConfig(), actor.GetServiceUrl()) // generic param is service url
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}
	pubKey, err := cryptoPocket.NewPublicKey(actor.GetPublicKey())
	if err != nil {
		return nil, err
	}

	peer := &typesP2P.NetworkPeer{
		Transport:  conn,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceURL: actor.GetServiceUrl(), // service url
	}

	return peer, nil
}

// WithConnectionFactory allows the user to specify a custom connection factory
func WithConnectionFactory(connFactory typesP2P.ConnectionFactory) func(PeerstoreProvider) {
	return func(ap PeerstoreProvider) {
		ap.SetConnectionFactory(connFactory)
	}
}
