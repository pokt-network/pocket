package types

import coreTypes "github.com/pokt-network/pocket/shared/core/types"

//go:generate mockgen -source=$GOFILE -destination=./mocks/addrbook_provider_mock.go github.com/pokt-network/pocket/p2p/types AddrBookProvider

// AddrBook is a way of representing NetworkPeer sets
type AddrBook []*NetworkPeer

// AddrBookMap maps p2p addresses to their respective NetworkPeer.
//
// Since maps cannot be sorted arbitrarily in Go, to achieve sorting, we need to rely on `addrList` which is a slice of addresses/strings and therefore we can sort it the way we want.
type AddrBookMap map[string]*NetworkPeer

// AddrBookProvider is an interface that provides AddrBook accessors
type AddrBookProvider interface {
	GetStakedAddrBookAtHeight(height uint64) (AddrBook, error)
	ActorsToAddrBook(actors map[string]coreTypes.Actor) (AddrBook, error)
	ActorToNetworkPeer(actor coreTypes.Actor) (*NetworkPeer, error)
}
