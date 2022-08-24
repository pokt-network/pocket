package p2p

import (
	"fmt"
	"log"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// CLEANUP(drewsky): These functions will turn into more of a "ActorToAddrBook" when we have a closer
// integration with utility.
func ValidatorMapToAddrBook(cfg *genesis.P2PConfig, validators map[string]*genesis.Actor) (typesP2P.AddrBook, error) {
	book := make(typesP2P.AddrBook, 0)
	for _, v := range validators {
		networkPeer, err := ValidatorToNetworkPeer(cfg, v)
		if err != nil {
			log.Println("[WARN] Error connecting to validator: ", err)
			continue
		}
		book = append(book, networkPeer)
	}
	return book, nil
}

// CLEANUP(drewsky): These functions will turn into more of a "ActorToAddrBook" when we have a closer
// integration with utility.
func ValidatorToNetworkPeer(cfg *genesis.P2PConfig, v *genesis.Actor) (*typesP2P.NetworkPeer, error) {
	conn, err := CreateDialer(cfg, v.GenericParam) // service url
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}

	pubKey, err := cryptoPocket.NewPublicKey(v.PublicKey)
	if err != nil {
		return nil, err
	}

	peer := &typesP2P.NetworkPeer{
		Dialer:     conn,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceUrl: v.GenericParam, // service url
	}

	return peer, nil
}
