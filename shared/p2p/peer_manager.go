package p2p

import (
	"sort"
	"sync"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/crypto"
)

// TECHDEBT(#553): can this be simplified / consolidated in to something else?
type PeersView interface {
	GetAddrs() []string
	GetPeers() PeerList
	GetPeerstore() Peerstore
}

type PeerManager interface {
	// TECHDEBT: move this to some new `ConnManager` interface.
	GetSelfAddr() crypto.Address
	GetPeersView() PeersView
	// HandleEvent synchronously reacts to `PeerManagerEvent`s
	HandleEvent(event PeerManagerEvent)
}

const (
	AddPeerEventType = PeerManagerEventType(iota)
	RemovePeerEventType
)

var _ PeerManager = &SortedPeerManager{}

type PeerManagerEventType int

type PeerManagerEvent struct {
	EventType PeerManagerEventType
	Peer      Peer
}

type SortedPeerManager struct {
	startAddr crypto.Address

	// eventCh provides a way for processing additions and removals from the pstore in an event-sourced way.
	//
	// The idea is that when we receive one of these events, we update the peersManager internal data structures
	// so that it can present a consistent state but in an optimized way.
	eventCh chan PeerManagerEvent

	m  sync.RWMutex
	wg sync.WaitGroup

	sortedPeers PeerList
	sortedAddrs []string
	pstore      Peerstore
}

type SortedPeersView struct {
	pstore      Peerstore
	sortedAddrs []string
	sortedPeers PeerList
}

// NewSortedPeerManager creates a new SortedPeerManager instance, it is
// responsible for handling operations on `sortedAddrs` and `sortedPeers`
// (like adding/removing them) within a Peerstore. It also takes care of
// keeping them sorted and indexed for fast access.
//
// If `isDynamic` is false, the peersManager will not handle addressBook changes,
// it will only be used for querying the PeerAddrMap
// TECHDEBT: signature should include a logger reference.
func NewSortedPeerManager(startAddr crypto.Address, pstore Peerstore, isDynamic bool) (*SortedPeerManager, error) {
	pm := &SortedPeerManager{
		startAddr:   startAddr,
		eventCh:     make(chan PeerManagerEvent, 1),
		pstore:      pstore,
		sortedAddrs: make([]string, pstore.Size()),
		sortedPeers: pstore.GetAllPeers(),
	}

	// initialize sortedAddrs
	for i, peer := range pstore.GetAllPeers() {
		pm.sortedAddrs[i] = peer.GetAddress().String()
	}
	sort.Strings(pm.sortedAddrs)

	i := sort.SearchStrings(pm.sortedAddrs, pm.startAddr.String())
	if i == len(pm.sortedAddrs) {
		logger.Global.Warn().
			Str("address", pm.startAddr.String()).
			Str("mode", "client-only").
			Msg("self address not found in peerstore so this client can send messages but does not propagate them")
	}
	// TECHDEBT: this message implies the `SortedPeerManager` knows too much about how it will be used.
	// Consider moving this check out to a "connection manager" in future refactoring.
	// Sorting is done lexicographically above, but is modified here so this addr of this node
	// is always the first in the list. This makes RainTree propagation easier to compute and interpret.
	pm.sortedAddrs = append(pm.sortedAddrs[i:len(pm.sortedAddrs)], pm.sortedAddrs[0:i]...)

	// sorting pm.sortedPeers as well, leveraging the sort order we just achieved
	for i := 0; i < len(pm.sortedAddrs); i++ {
		pm.sortedPeers[i] = pm.pstore.GetPeerFromString(pm.sortedAddrs[i])
	}

	if !isDynamic {
		return pm, nil
	}

	// listening and reacting to peer changes (addition/deletion) events
	go func() {
		for evt := range pm.eventCh {
			pm.m.Lock()

			peerAddress := evt.Peer.GetAddress()

			switch evt.EventType {
			case AddPeerEventType:
				if err := pm.pstore.AddPeer(evt.Peer); err != nil {
					logger.Global.Error().Err(err).
						Bool("TODO", true).
						Msgf("adding peer to peerstore")
				}
				// insert into sorted sortedAddrs and pstore
				// searching from index 1 because index 0 is self by convention and the rest of the slice is sorted
				i := sort.SearchStrings(pm.sortedAddrs[1:], peerAddress.String())
				pm.sortedAddrs = insertElementAtIndex(pm.sortedAddrs, peerAddress.String(), i)
				pm.sortedPeers = insertElementAtIndex(pm.sortedPeers, evt.Peer, i)

				pm.wg.Done()
			case RemovePeerEventType:
				if err := pm.pstore.RemovePeer(peerAddress); err != nil {
					logger.Global.Error().Err(err).
						Bool("TODO", true).
						Msgf("removing peer from peerstore")
				}

				// remove from sorted sortedAddrs and sortedPeers
				// searching from index 1 because index 0 is self by convention and the rest of the slice is sorted
				i := sort.SearchStrings(pm.sortedAddrs[1:], peerAddress.String())
				pm.sortedAddrs = removeElementAtIndex(pm.sortedAddrs, i)
				pm.sortedPeers = removeElementAtIndex(pm.sortedPeers, i)

				pm.wg.Done()
			}

			pm.m.Unlock()
		}
	}()

	return pm, nil
}

func (sortedPM *SortedPeerManager) GetPeersView() PeersView {
	sortedPM.m.RLock()
	defer sortedPM.m.RUnlock()

	// TECHDEBT: consider duplicating to avoid unintentional modification
	// by BasePeerManager consumers.
	return SortedPeersView{
		sortedPeers: sortedPM.sortedPeers,
		sortedAddrs: sortedPM.sortedAddrs,
		pstore:      sortedPM.pstore,
	}
}

func (sortedPM *SortedPeerManager) HandleEvent(event PeerManagerEvent) {
	sortedPM.wg.Add(1)
	sortedPM.eventCh <- event
	sortedPM.wg.Wait()
}

func (sortedPM *SortedPeerManager) GetSelfAddr() crypto.Address {
	return sortedPM.startAddr
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

func (view SortedPeersView) GetAddrs() []string {
	// TECHDEBT: consider freezing/duplicating to avoid unintentional modification.
	return view.sortedAddrs
}

func (view SortedPeersView) GetPeers() PeerList {
	// TECHDEBT: consider freezing/duplicating to avoid unintentional modification.
	return view.sortedPeers
}

func (view SortedPeersView) GetPeerstore() Peerstore {
	// TECHDEBT: consider freezing/duplicating to avoid unintentional modification.
	return view.pstore
}
