package types

// AddrBookMap maps p2p addresses to their respective NetworkPeer.
//
// Since maps cannot be sorted arbitrarily in Go, to achieve sorting, we need to rely on `addrList` which is a slice of addresses/strings and therefore we can sort it the way we want.
type AddrBookMap map[string]*NetworkPeer
