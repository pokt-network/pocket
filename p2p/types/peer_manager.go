package types

import (
	"sync"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/crypto"
)

type PeerManager interface {
	// TECHDEBT(#576): move this to some new `ConnManager` interface.
	GetSelfAddr() crypto.Address
	GetPeerstore() Peerstore
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
	// The idea is that when we receive one of these events, we update the view
	// to be sorted in the expected order and format without having to compute
	// it every time.
	eventCh chan PeerManagerEvent

	m  sync.RWMutex
	wg sync.WaitGroup

	view   *sortedPeersView
	pstore Peerstore
}

// NewSortedPeerManager creates a new SortedPeerManager instance, it is
// responsible for handling operations on `sortedAddrs` and `sortedPeers`
// (like adding/removing them) within a Peerstore. It also takes care of
// keeping them sorted and indexed for fast access.
//
// `startAddr` is intended to be that of the host/peer using the peer manager.
// It's kept at the beginning of the sorted lists in exception to the sorting for
// the convenience of the consumer. Used to identifying themselves, conventionally,
// within these primitive data structures.
//
// If `isDynamic` is false, the peersManager will not handle addressBook changes,
// it will only be used for querying the PeerAddrMap
// TECHDEBT: signature should include a logger reference.
func NewSortedPeerManager(startAddr crypto.Address, pstore Peerstore, isDynamic bool) (*SortedPeerManager, error) {
	pm := &SortedPeerManager{
		startAddr: startAddr,
		eventCh:   make(chan PeerManagerEvent, 1),
		pstore:    pstore,
		view:      NewSortedPeersView(startAddr, pstore),
	}

	if !isDynamic {
		return pm, nil
	}

	// TECKDEBT: moving this out to a "connection manager" in future refactoring.
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
				pm.view.Add(evt.Peer)
				pm.wg.Done()
			case RemovePeerEventType:
				if err := pm.pstore.RemovePeer(peerAddress); err != nil {
					logger.Global.Error().Err(err).
						Bool("TODO", true).
						Msgf("removing peer from peerstore")
				}

				pm.view.Remove(evt.Peer.GetAddress())
				pm.wg.Done()
			}

			pm.m.Unlock()
		}
	}()

	return pm, nil
}

func (sortedPM *SortedPeerManager) GetPeerstore() Peerstore {
	return sortedPM.pstore
}

func (sortedPM *SortedPeerManager) GetPeersView() PeersView {
	sortedPM.m.RLock()
	defer sortedPM.m.RUnlock()

	return sortedPM.view
}

func (sortedPM *SortedPeerManager) HandleEvent(event PeerManagerEvent) {
	sortedPM.wg.Add(1)
	sortedPM.eventCh <- event
	sortedPM.wg.Wait()
}

func (sortedPM *SortedPeerManager) GetSelfAddr() crypto.Address {
	return sortedPM.startAddr
}
