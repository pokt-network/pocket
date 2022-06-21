package types

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// TODO(olshansky): See if we can deprecate one of these structures.
type AddrBook []*NetworkPeer
type AddrBookMap map[string]*NetworkPeer

// TODO(olshansky): When we delete `stdnetwork` and only go with `raintree`, this interface
// can be simplified greatly.
type Network interface {
	NetworkBroadcast(data []byte) error
	NetworkSend(data []byte, address cryptoPocket.Address) error

	// Address book helpers
	GetAddrBook() AddrBook
	AddPeerToAddrBook(peer *NetworkPeer) error
	RemovePeerToAddrBook(peer *NetworkPeer) error

	// This function was added to support the raintree implementation.
	HandleNetworkData(data []byte) ([]byte, error)
}

type NetworkPeer struct {
	Dialer     TransportLayerConn
	PublicKey  cryptoPocket.PublicKey
	Address    cryptoPocket.Address
	ServiceUrl string // This is only included because it's a more human-friendly differentiator between peers
}

type TransportLayerConn interface {
	IsListener() bool
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}
