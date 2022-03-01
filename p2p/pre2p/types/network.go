package types

import (
	"net"

	pcrypto "github.com/pokt-network/pocket/shared/crypto"
)

type Network interface {
	NetworkBroadcast(data []byte) error
	NetworkSend(data []byte, address pcrypto.Address) error
	GetAddrBook() []*NetworkPeer
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr
	PublicKey     pcrypto.PublicKey
}
