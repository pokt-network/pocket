package types

import (
	"net"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

type AddrBook []*NetworkPeer

// TODO(olshansky): See which of these structures is better overall
type AddrBookMap map[string]*NetworkPeer

type Network interface {
	NetworkBroadcast(data []byte) error
	NetworkSend(data []byte, address cryptoPocket.Address) error
	GetAddrBook() AddrBook

	// TODO(olshansky): This should not be a separate interface from `NetworkBroadcast`
	// Similar to broadcast but when we are not the originator
	NetworkPropagate(msg *anypb.Any) error

	// TODO(olshansky): Discuss if we should just have an `Update` method or whether this should accept a list.
	AddPeerToAddrBook(peer *NetworkPeer) error
	RemovePeerToAddrBook(peer *NetworkPeer) error
}

type NetworkPeer struct {
	ConsensusAddr *net.TCPAddr
	PublicKey     cryptoPocket.PublicKey
	Address       cryptoPocket.Address
}
