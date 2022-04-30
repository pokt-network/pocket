package raintree

import (
	"fmt"
	"math"
	"sort"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// Refer to the specification for a formal description and proof of how the constants
// and math functions here were determined.
const (
	firstMsgTargetPercentage  = float64(1) / float64(3)
	secondMsgTargetPercentage = float64(2) / float64(3)
	shrinkagePercentage       = float64(2) / float64(3)
	maxLevelsLogBase          = float64(3)
)

// Whenever `addrBook` changes, we also need to update `addrBookMap` and `addrList`.
// TODO(olshansky): This is a very naive approach for now that recomputes everything every time that we can optimize later
func (n *rainTreeNetwork) handleAddrBookUpdates() error {
	n.addrBookMap = make(map[string]*typesPre2P.NetworkPeer, len(n.addrBook))
	n.addrList = make([]string, len(n.addrBook))
	for i, peer := range n.addrBook {
		addr := peer.Address.String()
		n.addrList[i] = addr
		n.addrBookMap[addr] = peer
	}
	n.maxNumLevels = n.getMaxAddrBookLevels()

	sort.Strings(n.addrList)
	if i, ok := n.getSelfIndexInAddrBook(); !ok {
		return fmt.Errorf("self address not found in addrBook so this client can send messages but does not propagate them")
	} else {
		// We sorted the list lexicographically above, but we reformat it here so
		// the address of this node is first in the list to make RainTree propagation
		// easier to compute and interpret.
		n.addrList = append(n.addrList[i:len(n.addrList)], n.addrList[0:i]...)
	}

	return nil
}

func (n *rainTreeNetwork) getFirstTargetAddr(level uint32) (cryptoPocket.Address, bool) {
	l := n.getAddrBookLengthAtHeight(level)
	i := int(firstMsgTargetPercentage * float64(l))
	addrStr := n.addrList[i]
	return n.addrBookMap[addrStr].Address, true
}

func (n *rainTreeNetwork) getSecondTargetAddr(level uint32) (cryptoPocket.Address, bool) {
	l := n.getAddrBookLengthAtHeight(level)
	i := int(firstMsgTargetPercentage * float64(l))
	addrStr := n.addrList[i]
	return n.addrBookMap[addrStr].Address, true
}

func (n *rainTreeNetwork) getSelfIndexInAddrBook() (int, bool) {
	addrString := n.addr.String()
	for i, addr := range n.addrList {
		if addr == addrString {
			return i, true
		}
	}
	return -1, false
}

// TODO(drewsky): Could we hit an issue where we are propagating a message from an older height
// (e.g. before the addr book was updated), but we're using `maxNumLevels` associated with the
// current height.
func (n *rainTreeNetwork) getAddrBookLengthAtHeight(level uint32) int {
	shrinkageCoefficient := math.Pow(shrinkagePercentage, float64(n.maxNumLevels-level))
	return int(math.Ceil(float64(len(n.addrList)) * shrinkageCoefficient))
}

func (n *rainTreeNetwork) getMaxAddrBookLevels() uint32 {
	addrBookSize := float64(len(n.addrBook))
	return uint32(math.Ceil(logBase(addrBookSize)*100) / 100)
}

func logBase(x float64) float64 {
	return math.Log(x) / math.Log(maxLevelsLogBase)
}
