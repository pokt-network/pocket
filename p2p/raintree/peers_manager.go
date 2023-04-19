package raintree

import (
	"math"
	"sort"
	"sync"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

var _ typesP2P.PeerManager = &rainTreePeersManager{}

// rainTreePeersManager is in charge of handling operations on peers (like adding/removing them) within an Peerstore
type rainTreePeersManager struct {
	*typesP2P.SortedPeerManager

	maxLevelsMutex sync.Mutex
	maxNumLevels   uint32
}

func newPeersManagerWithPeerstoreProvider(selfAddr cryptoPocket.Address, pstoreProvider peerstore_provider.PeerstoreProvider, height uint64) (*rainTreePeersManager, error) {
	pstore, err := pstoreProvider.GetStakedPeerstoreAtHeight(height)
	if err != nil {
		return nil, err
	}

	return newPeersManager(selfAddr, pstore, false)
}

// newPeersManager creates a new rainTreePeersManager instance, it is in charge of handling operations on peers (like adding/removing them) within an Peerstore
// it also takes care of keeping the Peerstore sorted and indexed for fast access
//
// If `isDynamic` is false, the rainTreePeersManager will not handle addressBook changes, it will only be used for querying the Peerstore
func newPeersManager(selfAddr cryptoPocket.Address, pstore typesP2P.Peerstore, isDynamic bool) (*rainTreePeersManager, error) {
	sortedPM, err := typesP2P.NewSortedPeerManager(selfAddr, pstore, isDynamic)
	if err != nil {
		return nil, err
	}

	pm := &rainTreePeersManager{
		SortedPeerManager: sortedPM,
		maxNumLevels:      0,
	}

	// initializing map and list
	pm.maxNumLevels = pm.getMaxPeerstoreLevels()

	return pm, nil
}

func (pm *rainTreePeersManager) HandleEvent(evt typesP2P.PeerManagerEvent) {
	pm.SortedPeerManager.HandleEvent(evt)
	pm.updateMaxNumLevels()
}

func (pm *rainTreePeersManager) GetPeersView() typesP2P.PeersView {
	return pm.SortedPeerManager.GetPeersView()
}

func (pm *rainTreePeersManager) GetMaxNumLevels() uint32 {
	pm.maxLevelsMutex.Lock()
	defer pm.maxLevelsMutex.Unlock()

	return pm.maxNumLevels
}

func (pm *rainTreePeersManager) getPeersViewWithLevels() (view typesP2P.PeersView, level uint32) {
	return pm.GetPeersView(), pm.GetMaxNumLevels()
}

// DISCUSS: This is only used in tests. Should we remove it?
func (pm *rainTreePeersManager) getSelfIndexInPeersView() (int, bool) {
	addrs := pm.GetPeersView().GetAddrs()
	if len(addrs) == 0 {
		return -1, false
	}
	if addrs[0] == pm.SortedPeerManager.GetSelfAddr().String() {
		return 0, true
	}
	i := sort.SearchStrings(addrs[1:], pm.GetSelfAddr().String())
	if i == len(addrs) {
		return -1, false
	}
	return i, true
}

func (pm *rainTreePeersManager) getMaxPeerstoreLevels() uint32 {
	pstoreSize := pm.GetPeerstore().Size()
	return uint32(math.Ceil(logBase(float64(pstoreSize))))
}

func (pm *rainTreePeersManager) updateMaxNumLevels() {
	pm.maxNumLevels = pm.getMaxPeerstoreLevels()
}

func logBase(x float64) float64 {
	return round(math.Log(x)/math.Log(maxLevelsLogBase), floatPrecision)
}

func round(value, precision float64) float64 {
	return math.Round(value/precision) * precision
}
