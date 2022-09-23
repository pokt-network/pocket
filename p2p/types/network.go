package types

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// CLEANUP(olshansky): See if we can deprecate one of these structures.
type AddrBook []*NetworkPeer
type AddrList []string
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

type NetworkPeer struct {
	Dialer     Transport
	PublicKey  cryptoPocket.PublicKey
	Address    cryptoPocket.Address
	ServiceUrl string // This is only included because it's a more human-friendly differentiator between peers
}

type Transport interface {
	IsListener() bool
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}

// TODO (Team) break off these types into diff files

func (ab *AddrList) Find(address string) (index int, found bool) {
	if ab == nil {
		return 0, false
	}
	addressBook := *ab
	for i, a := range addressBook {
		if a == address {
			return i, true
		}
	}
	return 0, false
}
