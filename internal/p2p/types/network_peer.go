package types

import (
	cryptoPocket "github.com/pokt-network/pocket/internal/shared/crypto"
)

type NetworkPeer struct {
	Dialer    Transport
	PublicKey cryptoPocket.PublicKey
	Address   cryptoPocket.Address

	// This is only included because it's a more human-friendly differentiator between peers
	ServiceUrl string
}
