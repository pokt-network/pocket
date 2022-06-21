package pre2p

import (
	"fmt"
	"log"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

// CLEANUP(drewsky): These functions will turn into more of a "ActorToAddrBook" when we have a closer
// integration with utility.
func ValidatorMapToAddrBook(cfg *config.Pre2PConfig, validators map[string]*typesGenesis.Validator) (typesPre2P.AddrBook, error) {
	book := make(typesPre2P.AddrBook, 0)
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
func ValidatorToNetworkPeer(cfg *config.Pre2PConfig, v *typesGenesis.Validator) (*typesPre2P.NetworkPeer, error) {
	conn, err := CreateDialer(cfg, v.ServiceUrl)
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}

	pubKey, err := cryptoPocket.NewPublicKeyFromBytes(v.PublicKey)
	if err != nil {
		return nil, err
	}

	peer := &typesPre2P.NetworkPeer{
		Dialer:     conn,
		PublicKey:  pubKey,
		Address:    pubKey.Address(),
		ServiceUrl: v.ServiceUrl,
	}

	return peer, nil
}
