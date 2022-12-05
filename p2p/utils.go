package p2p

import (
	"fmt"
	"log"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

func ActorToAddrBook(cfg modules.P2PConfig, validators map[string]modules.Actor) (typesP2P.AddrBook, error) {
	book := make(typesP2P.AddrBook, 0)
	for _, v := range validators {
		networkPeer, err := ActorToNetworkPeer(cfg, v)
		if err != nil {
			log.Println("[WARN] Error connecting to validator: ", err)
			continue
		}
		book = append(book, networkPeer)
	}
	return book, nil
}

func ActorToNetworkPeer(cfg modules.P2PConfig, v modules.Actor) (*typesP2P.NetworkPeer, error) {
	conn, err := CreateDialer(cfg, v.GetGenericParam()) // service url
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}

	pubKey, err := cryptoPocket.NewPublicKey(v.GetPublicKey())
	if err != nil {
		return nil, err
	}

	peer := &typesP2P.NetworkPeer{
		Dialer:     conn,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceUrl: v.GetGenericParam(), // service url
	}

	return peer, nil
}
