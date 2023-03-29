package types

import (
	"sort"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/crypto"
)

// CONSIDERATION(#576): does it make more sense to move this interface to the
// consumer package and unexport?
// Pro: interface can be unexported
// Con: can't make interface assignment compile-time check
type PeersView interface {
	GetAddrs() []string
	GetPeers() PeerList
}

var _ PeersView = &sortedPeersView{}

type sortedPeersView struct {
	sortedAddrs []string
	sortedPeers PeerList
}

// NewSortedPeersView constructs a `sortedPeersView` from the given `Peerstore`.
// `startAddr` is kept at the first index of `sortedAddrs` by convention, same
// for the respective `Peer` in `sortedPeers`. See `NewSortedPeerManager` for more.
func NewSortedPeersView(startAddr crypto.Address, pstore Peerstore) *sortedPeersView {
	view := &sortedPeersView{
		sortedAddrs: make([]string, pstore.Size()),
		sortedPeers: pstore.GetPeerList(),
	}
	return view.init(startAddr, pstore)
}

func insertElementAtIndex[T any](slice []T, element T, index int) []T {
	slice = append(slice, element)
	copy(slice[index+1:], slice[index:])
	slice[index] = element
	return slice
}

func removeElementAtIndex[T any](slice []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, slice[:index]...)
	return append(ret, slice[index+1:]...)
}

// GetAddrs implements the respective member of the `PeersView` interface.
func (view *sortedPeersView) GetAddrs() []string {
	return view.sortedAddrs
}

// GetPeers implements the respective member of the `PeersView` interface.
func (view *sortedPeersView) GetPeers() PeerList {
	return view.sortedPeers
}

// Add inserts into sorted sortedAddrs and sortedPeers.
// Searches from index 1 because index 0 is self by convention and the rest of
// the slice is sorted.
func (view *sortedPeersView) Add(peer Peer) {
	i := view.getAddrIndex(peer.GetAddress())

	view.sortedAddrs = insertElementAtIndex(view.sortedAddrs, peer.GetAddress().String(), i)
	view.sortedPeers = insertElementAtIndex(view.sortedPeers, peer, i)
}

// Remove removes the peer with the given address from sortedAddrs and sortedPeers.
// Searches from index 1 because index 0 is self by convention and the rest of
// the slice is sorted.
func (view *sortedPeersView) Remove(addr crypto.Address) {
	i := view.getAddrIndex(addr)
	if i == len(view.sortedAddrs) {
		logger.Global.Debug().
			Str("pokt_addr", addr.String()).
			Msg("not found in view.sortedAddrs")
		return
	}

	view.sortedAddrs = removeElementAtIndex(view.sortedAddrs, i)
	view.sortedPeers = removeElementAtIndex(view.sortedPeers, i)
}

// init copies peers and addresses from `pstore` into `sortedAddrs` and
// `sortedPeers`, and then sorts both. Returns itself for convenience.
func (view *sortedPeersView) init(startAddr crypto.Address, pstore Peerstore) *sortedPeersView {
	for i, peer := range pstore.GetPeerList() {
		view.sortedAddrs[i] = peer.GetAddress().String()
	}

	// Copying sortedPeers, preserving the sort order of sortedAddrs.
	for i := 0; i < len(view.sortedAddrs); i++ {
		view.sortedPeers[i] = pstore.GetPeerFromString(view.sortedAddrs[i])
	}

	view.sortAddrs(startAddr)
	return view
}

// sortAddrs sorts addresses in `sortedAddrs` lexicographically then shifts
// `startAddr` to the first index, moving any preceding values to the end
// of the list; effectively preserving the order by "wrapping around".
func (view *sortedPeersView) sortAddrs(startAddr crypto.Address) {
	sort.Strings(view.sortedAddrs)

	i := sort.SearchStrings(view.sortedAddrs, startAddr.String())
	if i == len(view.sortedAddrs) {
		logger.Global.Warn().
			Str("address", startAddr.String()).
			Str("mode", "client-only").
			Msg("self address not found in peerstore so this client can send messages but does not propagate them")
	}
	view.sortedAddrs = append(view.sortedAddrs[i:len(view.sortedAddrs)], view.sortedAddrs[0:i]...)
}

// getAddrIndex returns the sortedAddrs index at which the given address is stored
// or at which to insert it if not present.
func (view *sortedPeersView) getAddrIndex(addr crypto.Address) int {
	wrapIdx := sort.Search(len(view.sortedAddrs), func(visitIdx int) bool {
		return view.sortedAddrs[visitIdx] < view.sortedAddrs[0]
	})

	frontAddrs := view.sortedAddrs[:wrapIdx]
	backAddrs := view.sortedAddrs[wrapIdx:]
	i := sort.SearchStrings(frontAddrs, addr.String())
	if i == 0 {
		i = sort.SearchStrings(backAddrs, addr.String())
		i += len(frontAddrs)
	}
	return i
}
