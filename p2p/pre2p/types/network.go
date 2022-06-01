package types

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// CLEANUP(olshansky): See if we can deprecate one of these structures.
type AddrBook []*NetworkPeer
type AddrBookMap map[string]*NetworkPeer

// TECHDEBT(olshansky): When we delete `stdnetwork` and only go with `raintree`, this interface
// can be simplified greatly.
type Network interface {
	NetworkBroadcast(data []byte) error
	NetworkSend(data []byte, address cryptoPocket.Address) error

	// Address book helpers
	GetAddrBook() AddrBook
	AddPeerToAddrBook(peer *NetworkPeer) error    // TODO(team): Not used yet
	RemovePeerToAddrBook(peer *NetworkPeer) error // TODO(team): Not used yet

	// This function was added to specifically support the RainTree implementation.
	// Handles the raw data received from the network and returns the data to be processed
	// by the application layer.
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
