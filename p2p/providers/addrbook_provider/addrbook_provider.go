package addrbook_provider

//go:generate mockgen -source=$GOFILE -destination=../../types/mocks/addrbook_provider_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types AddrBookProvider

import (
	"fmt"
	"log"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// AddrBookProvider is an interface that provides AddrBook accessors
type AddrBookProvider interface {
	GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error)
	GetConnFactory() typesP2P.ConnectionFactory
	GetP2PConfig() modules.P2PConfig
	SetConnectionFactory(typesP2P.ConnectionFactory)
}

func ActorsToAddrBook(abp AddrBookProvider, actors []modules.Actor) (typesP2P.AddrBook, error) {
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

func ActorToNetworkPeer(abp AddrBookProvider, actor modules.Actor) (*typesP2P.NetworkPeer, error) {
	conn, err := abp.GetConnFactory()(abp.GetP2PConfig(), actor.GetGenericParam()) // service url
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}

	pubKey, err := cryptoPocket.NewPublicKey(actor.GetPublicKey())
	if err != nil {
		return nil, err
	}

	peer := &typesP2P.NetworkPeer{
		Dialer:     conn,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceUrl: actor.GetGenericParam(), // service url
	}

	return peer, nil
}

// WithConnectionFactory allows the user to specify a custom connection factory
func WithConnectionFactory(connFactory typesP2P.ConnectionFactory) func(AddrBookProvider) {
	return func(ap AddrBookProvider) {
		ap.SetConnectionFactory(connFactory)
	}
}
