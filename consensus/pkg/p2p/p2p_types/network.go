package p2p_types

import (
	"crypto"
	"net"

	"pocket/consensus/pkg/types"
)

type Network interface {
	ConnectToValidator(nodeId types.NodeId, v *types.Validator) error
	NetworkBroadcast(data []byte, self types.NodeId) error
	NetworkSend(data []byte, node types.NodeId) error
	GetAddrBook() []*NetworkPeer
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr
	DebugAddr     *net.TCPAddr

	NodeId    types.NodeId
	PublicKey crypto.PublicKey
}
