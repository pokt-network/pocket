package raintree

import (
	"log"
	"math"
	"sort"
	"sync"

	"github.com/pokt-network/pocket/p2p/types"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// peersManager is in charge of handling operations on peers (like adding/removing them) within an AddrBook
type peersManager struct {
	selfAddr cryptoPocket.Address

	// eventCh provides a way for processing additions and removals from the addrBook in an event-sourced way.
	//
	// The idea is that when we receive one of these events, we update the peersManager internal data structures
	// so that it can present a consistent state but in an optimized way.
	eventCh chan addressBookEvent

	m  sync.RWMutex
	wg sync.WaitGroup

	addrBook     typesP2P.AddrBook
	addrBookMap  typesP2P.AddrBookMap
	addrList     typesP2P.AddrList
	maxNumLevels uint32
}

func NewPeersManagerWithAddrBookProvider(selfAddr cryptoPocket.Address, addrBookProvider typesP2P.AddrBookProvider, height uint64) (*peersManager, error) {
	addrBook, err := addrBookProvider(height)
	if err != nil {
		return nil, err
	}

	return newPeersManager(selfAddr, addrBook, false)
}

// newPeersManager creates a new peersManager instance, it is in charge of handling operations on peers (like adding/removing them) within an AddrBook
// it also takes care of keeping the AddrBook sorted and indexed for fast access
//
// If `isDynamic` is true, the peersManager will not handle addressBook changes, it will only be used for querying the AddrBook
func newPeersManager(selfAddr cryptoPocket.Address, addrBook typesP2P.AddrBook, isDynamic bool) (*peersManager, error) {
	pm := &peersManager{
		selfAddr:     selfAddr,
		addrBook:     addrBook,
		eventCh:      make(chan addressBookEvent, 1),
		addrBookMap:  make(typesP2P.AddrBookMap),
		addrList:     make([]string, 0),
		maxNumLevels: 0,
	}

	// initializing map and list
	pm.addrBookMap = make(map[string]*typesP2P.NetworkPeer, len(addrBook))
	pm.addrList = make([]string, len(addrBook))
	for i, peer := range addrBook {
		addr := peer.Address.String()
		pm.addrList[i] = addr
		pm.addrBookMap[addr] = peer
	}

	sort.Strings(pm.addrList)

	i := sort.SearchStrings(pm.addrList, pm.selfAddr.String())
	if i == len(pm.addrList) {
		log.Printf("[⚠️ client-only mode]: self address not found for %s in addrBook so this client can send messages but does not propagate them", pm.selfAddr)
	}
	// The list is sorted lexicographically above, but is reformatted below so this addr of this node
	// is always the first in the list. This makes RainTree propagation easier to compute and interpret.
	pm.addrList = append(pm.addrList[i:len(pm.addrList)], pm.addrList[0:i]...)

	// sorting pm.addrBook as well, leveraging the sort order we just achieved
	for i := 0; i < len(pm.addrList); i++ {
		pm.addrBook[i] = pm.addrBookMap[pm.addrList[i]]
	}

	pm.maxNumLevels = pm.getMaxAddrBookLevels()

	if !isDynamic {
		return pm, nil
	}

	// listening and reacting to peer changes (addition/deletion) events
	go func() {
		for evt := range pm.eventCh {
			pm.m.Lock()

			peerAddress := evt.peer.Address.String()

			switch evt.eventType {
			case addToAddressBook:
				pm.addrBookMap[peerAddress] = evt.peer
				// insert into sorted addrList and addrBook
				// searching from index 1 because index 0 is self by convention and the rest of the slice is sorted
				i := sort.SearchStrings(pm.addrList[1:], peerAddress)
				pm.addrList = insertElementAtIndex(pm.addrList, peerAddress, i)
				pm.addrBook = insertElementAtIndex(pm.addrBook, evt.peer, i)

				updateMaxNumLevels(pm)

				pm.wg.Done()
			case removeFromAddressBook:
				delete(pm.addrBookMap, peerAddress)

				// remove from sorted addrList and addrBook
				// searching from index 1 because index 0 is self by convention and the rest of the slice is sorted
				i := sort.SearchStrings(pm.addrList[1:], peerAddress)
				pm.addrList = removeElementAtIndex(pm.addrList, i)
				pm.addrBook = removeElementAtIndex(pm.addrBook, i)

				updateMaxNumLevels(pm)

				pm.wg.Done()
			}

			pm.m.Unlock()
		}
	}()

	return pm, nil
}

func (pm *peersManager) getNetworkView() networkView {
	pm.m.RLock()
	defer pm.m.RUnlock()
	return networkView{
		addrBook:     pm.addrBook,
		addrBookMap:  pm.addrBookMap,
		addrList:     pm.addrList,
		maxNumLevels: pm.maxNumLevels,
	}
}

// DISCUSS: This is only used in tests. Should we remove it?
func (pm *peersManager) getSelfIndexInAddrBook() (int, bool) {
	if len(pm.addrList) == 0 {
		return -1, false
	}
	if pm.addrList[0] == pm.selfAddr.String() {
		return 0, true
	}
	i := sort.SearchStrings(pm.addrList[1:], pm.selfAddr.String())
	if i == len(pm.addrList) {
		return -1, false
	}
	return i, true
}

func (pm *peersManager) getMaxAddrBookLevels() uint32 {
	peersManagerStateView := pm.getNetworkView()
	addrBookSize := float64(len(peersManagerStateView.addrBookMap))
	return uint32(math.Ceil(logBase(addrBookSize)))
}

func logBase(x float64) float64 {
	return round(math.Log(x)/math.Log(maxLevelsLogBase), floatPrecision)
}

func round(value, precision float64) float64 {
	return math.Round(value/precision) * precision
}

type addressBookEventType bool

const (
	addToAddressBook      addressBookEventType = true
	removeFromAddressBook addressBookEventType = false
)

type addressBookEvent struct {
	eventType addressBookEventType
	peer      *types.NetworkPeer
}

type networkView struct {
	addrBook     typesP2P.AddrBook
	addrBookMap  typesP2P.AddrBookMap
	addrList     typesP2P.AddrList
	maxNumLevels uint32
}

func updateMaxNumLevels(pm *peersManager) {
	addrBookSize := float64(len(pm.addrBook))
	pm.maxNumLevels = uint32(math.Ceil(logBase(addrBookSize)))
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
