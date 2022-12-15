package types

//go:generate mockgen -source=$GOFILE -destination=./mocks/addrbook_provider_mock.go github.com/pokt-network/pocket/p2p/types AddrBookProvider

import "github.com/pokt-network/pocket/shared/modules"

// AddrBook is a way of representing NetworkPeer sets
type AddrBook []*NetworkPeer

// AddrBookMap maps p2p addresses to their respective NetworkPeer.
//
// Since maps cannot be sorted arbitrarily in Go, to achieve sorting, we need to rely on `addrList` which is a slice of addresses/strings and therefore we can sort it the way we want.
type AddrBookMap map[string]*NetworkPeer

// AddrBookProvider is an interface that provides AddrBook accessors
type AddrBookProvider interface {
	GetStakedAddrBookAtHeight(height uint64) (AddrBook, error)
	ValidatorMapToAddrBook(validators map[string]modules.Actor) (AddrBook, error)
	ValidatorToNetworkPeer(v modules.Actor) (*NetworkPeer, error)
}
