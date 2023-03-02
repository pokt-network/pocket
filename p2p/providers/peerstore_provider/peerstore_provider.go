package peerstore_provider

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/peerstore_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types PeerstoreProvider

import (
	"fmt"
	"strings"

	"github.com/pokt-network/pocket/logger"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
)

const ModuleName = "peerstore_provider"

// PeerstoreProvider is an interface that provides Peerstore accessors
type PeerstoreProvider interface {
	modules.Module

	GetStakedPeerstoreAtHeight(height uint64) (sharedP2P.Peerstore, error)
	GetConnFactory() typesP2P.ConnectionFactory
	GetP2PConfig() *configs.P2PConfig
	SetConnectionFactory(typesP2P.ConnectionFactory)
}

func ActorsToPeerstore(abp PeerstoreProvider, actors []*coreTypes.Actor) (sharedP2P.Peerstore, error) {
	// TECHDEBT: consider using a multi-error pkg or upgrading to go 1.20.
	// (see: https://go.dev/doc/go1.20#errors)
	var errs []string
	appendErr := func(err error) {
		errs = append(errs, err.Error())
	}
	joinErrs := func() string {
		return strings.Join(errs, "; ")
	}

	pstore := make(sharedP2P.PeerAddrMap)
	for _, a := range actors {
		networkPeer, err := ActorToPeer(abp, a)
		if err != nil {
			appendErr(err)
			continue
		}

		if err = pstore.AddPeer(networkPeer); err != nil {
			appendErr(err)
		}

		// TECHDEBT: consider using a multi-error and returning instead of logging.
		if len(errs) > 0 {
			logger.Global.Warn().
				Bool("TODO", true).
				Msgf("building peerstore from actors list: %s", joinErrs())
		}
	}
	return pstore, nil
}

func ActorToPeer(abp PeerstoreProvider, actor *coreTypes.Actor) (sharedP2P.Peer, error) {
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
