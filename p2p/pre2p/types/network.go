package types

import (
	"crypto"
	"net"
)

// TODO(olshansky): Try to find a way to remove `NodeId` from the entire codebase altogether; this is a stop-gap prototype measure.
type Network interface {
	NetworkBroadcast(data []byte, self NodeId) error
	NetworkSend(data []byte, node NodeId) error
	GetAddrBook() []*NetworkPeer
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr

	NodeId    NodeId
	PublicKey crypto.PublicKey
}
