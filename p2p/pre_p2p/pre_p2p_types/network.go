package pre_p2p_types

import (
	"crypto"
	"net"
)

type Network interface {
	ConnectToValidator(nodeId NodeId, v *Validator) error
	NetworkBroadcast(data []byte, self NodeId) error
	NetworkSend(data []byte, node NodeId) error
	GetAddrBook() []*NetworkPeer
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr
	DebugAddr     *net.TCPAddr

	NodeId    NodeId
	PublicKey crypto.PublicKey
}
