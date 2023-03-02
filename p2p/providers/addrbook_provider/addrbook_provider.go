package addrbook_provider

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/addrbook_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types AddrBookProvider

import (
	"fmt"
	"log"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

const ModuleName = "addrbook_provider"

// AddrBookProvider is an interface that provides AddrBook accessors
type AddrBookProvider interface {
	modules.Module

	GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error)
	GetConnFactory() typesP2P.ConnectionFactory
	GetP2PConfig() *configs.P2PConfig
	SetConnectionFactory(typesP2P.ConnectionFactory)
}

func ActorsToAddrBook(abp AddrBookProvider, actors []*coreTypes.Actor) (typesP2P.AddrBook, error) {
	book := make(typesP2P.AddrBook, 0)
	for _, a := range actors {
		networkPeer, err := ActorToNetworkPeer(abp, a)
		if err != nil {
			log.Println("[WARN] Error connecting to validator: ", err)
			continue
		}
		book = append(book, networkPeer)
	}
	return book, nil
}

func ActorToNetworkPeer(abp AddrBookProvider, actor *coreTypes.Actor) (*typesP2P.NetworkPeer, error) {
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
func WithConnectionFactory(connFactory typesP2P.ConnectionFactory) func(AddrBookProvider) {
	return func(ap AddrBookProvider) {
		ap.SetConnectionFactory(connFactory)
	}
}
