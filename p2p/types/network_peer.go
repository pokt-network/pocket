package types

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type NetworkPeer struct {
	Dialer     Transport
	PublicKey  cryptoPocket.PublicKey
	Address    cryptoPocket.Address
	ServiceUrl string // This is only included because it's a more human-friendly differentiator between peers
}
