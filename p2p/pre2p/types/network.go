package types

import (
	"net"
	pcrypto "pocket/shared/crypto"
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
