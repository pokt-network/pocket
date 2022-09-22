package types

//go:generate mockgen -source=$GOFILE -destination=./mocks/network_mock.go github.com/pokt-network/pocket/p2p/types Network

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// CLEANUP(olshansky): See if we can deprecate one of these structures.
// type AddrBook []*NetworkPeer
// type AddrList []string
type AddrBookMap map[string]*NetworkPeer

// TECHDEBT(olshansky): When we delete `stdnetwork` and only go with `raintree`, this interface
// can be simplified greatly.
type Network interface {
	modules.IntegratableModule

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
