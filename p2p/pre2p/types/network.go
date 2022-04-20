package types

import (
	"net"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

type Network interface {
	NetworkBroadcast(data []byte) error
	NetworkSend(data []byte, address cryptoPocket.Address) error
	GetAddrBook() []*NetworkPeer
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr
	PublicKey     cryptoPocket.PublicKey
}
