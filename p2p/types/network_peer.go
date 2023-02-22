package types

import (
	"github.com/multiformats/go-multiaddr"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type NetworkPeer struct {
	Dialer    Transport
	PublicKey cryptoPocket.PublicKey
	Address   cryptoPocket.Address
	Multiaddr multiaddr.Multiaddr

	// This is only included because it's a more human-friendly differentiator between peers
	ServiceUrl string
}
