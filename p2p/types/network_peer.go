package types

import (
	"fmt"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type NetworkPeer struct {
	Dialer    Transport
	PublicKey cryptoPocket.PublicKey
	Address   cryptoPocket.Address

	// This is only included because it's a more human-friendly differentiator between peers
	ServiceUrl string
}

func (peer *NetworkPeer) String() string {
	return fmt.Sprintf("address: %s, serviceURL: %s", peer.Address.String(), peer.ServiceUrl)
}
