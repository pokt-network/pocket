package types

import (
	"crypto"
	"net"
)

// TODO(olshansky): Try to find a way to remove `NodeId` from the entire codebase altogether. This is a stop-gap prototype measure.
type Network interface {
	NetworkBroadcast(data []byte, self NodeId) error
	NetworkSend(data []byte, node NodeId) error

	ConnectToValidator(nodeId NodeId, v *Validator) error

	GetAddrBook() []*NetworkPeer
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr
	DebugAddr     *net.TCPAddr

	NodeId    NodeId
	PublicKey crypto.PublicKey
}
