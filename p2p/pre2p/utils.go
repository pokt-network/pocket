package pre2p

import (
	"fmt"
	"log"
	"net"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
)

func ValidatorMapToAddrBook(validators map[string]*typesGenesis.Validator) (typesPre2P.AddrBook, error) {
	book := make(typesPre2P.AddrBook, 0)
	for _, v := range validators {
		networkPeer, err := ValidatorToNetworkPeer(v)
		if err != nil {
			log.Println("[WARN] Error connecting to validator: ", err)
			continue
		}
		book = append(book, networkPeer)
	}
	return book, nil
}

func ValidatorToNetworkPeer(v *typesGenesis.Validator) (*typesPre2P.NetworkPeer, error) {
	tcpAddr, err := net.ResolveTCPAddr(NetworkProtocol, v.ServiceUrl)
	if err != nil {
		return nil, fmt.Errorf("error resolving addr: %v", err)
	}

	pubKey, err := cryptoPocket.NewPublicKeyFromBytes(v.PublicKey)
	if err != nil {
		return nil, err
	}

	peer := &typesPre2P.NetworkPeer{
		ConsensusAddr: tcpAddr,
		PublicKey:     pubKey,
		Address:       pubKey.Address(),
	}

	return peer, nil
}
