package raintree

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// Refer to the P2P specification for a formal description and proof of how the constants are selected
const (
	firstMsgTargetPercentage  = float64(1) / float64(3)
	secondMsgTargetPercentage = float64(2) / float64(3)
	shrinkagePercentage       = float64(2) / float64(3)
	maxLevelsLogBase          = float64(3)
	floatPrecision            = float64(0.0000001)
)

// Whenever `addrBook` changes, we also need to update `addrBookMap` and `addrList`
func (n *rainTreeNetwork) handleAddrBookUpdates() error {
	// OPTIMIZE(olshansky): This is a very naive approach for now that recomputes everything every time that we can optimize later
	n.addrBookMap = make(map[string]*typesPre2P.NetworkPeer, len(n.addrBook))
	n.addrList = make([]string, len(n.addrBook))
	for i, peer := range n.addrBook {
		addr := peer.Address.String()
		n.addrList[i] = addr
		n.addrBookMap[addr] = peer
	}
	n.maxNumLevels = n.getMaxAddrBookLevels()

	sort.Strings(n.addrList)
	if i, ok := n.getSelfIndexInAddrBook(); ok {
		// The list is sorted lexicographically above, but is reformatted below so this addr of this node
		// is always the first in the list. This makes RainTree propagation easier to compute and interpret.
		n.addrList = append(n.addrList[i:len(n.addrList)], n.addrList[0:i]...)
	} else {
		return fmt.Errorf("self address not found for %s in addrBook so this client can send messages but does not propagate them", n.selfAddr)
	}
	return nil
}

func (n *rainTreeNetwork) getSelfIndexInAddrBook() (int, bool) {
	addrString := n.selfAddr.String()
	for i, addr := range n.addrList {
		if addr == addrString {
			return i, true
		}
	}
	return -1, false
}

func (n *rainTreeNetwork) getFirstTargetAddr(level uint32) (cryptoPocket.Address, bool) {
	return n.getTarget(level, firstMsgTargetPercentage)
}

func (n *rainTreeNetwork) getSecondTargetAddr(level uint32) (cryptoPocket.Address, bool) {
	return n.getTarget(level, secondMsgTargetPercentage)
}

func (n *rainTreeNetwork) getTarget(level uint32, targetPercentage float64) (cryptoPocket.Address, bool) {
	// OPTIMIZE(olshansky): We are computing this twice for each message, but it's not that expensive.
	l := n.getAddrBookLengthAtHeight(level)

	i := int(targetPercentage * float64(l))

	// If the target is 0, it is a reference to self, which is a `Demote` in RainTree terms.
	// This is handled separately.
	if i == 0 {
		return nil, false
	}

	addrStr := n.addrList[i]
	if addr, ok := n.addrBookMap[addrStr]; ok {
		// IMPROVE(olshansky): Consolidate so the debug print contains all (i.e. both) targets in one log line
		log.Printf("[DEBUG] Target (%0.2f) at height (%d): %s", targetPercentage, level, n.debugMsgTargetString(l, i))
		return addr.Address, true
	}
	return nil, false
}

// TODO(team): Need to integrate with persistence layer so we are storing this on a per height basis.
// We can easily hit an issue where we are propagating a message from an older height (e.g. before
// the addr book was updated), but we're using `maxNumLevels` associated with the number of
// validators at the current height.
func (n *rainTreeNetwork) getAddrBookLengthAtHeight(level uint32) int {
	shrinkageCoefficient := math.Pow(shrinkagePercentage, float64(n.maxNumLevels-level))
	return int(float64(len(n.addrList)) * (shrinkageCoefficient))
}

func (n *rainTreeNetwork) getMaxAddrBookLevels() uint32 {
	addrBookSize := float64(len(n.addrBook))
	return uint32(math.Ceil(logBase(addrBookSize)))
}

func logBase(x float64) float64 {
	return round(math.Log(x)/math.Log(maxLevelsLogBase), floatPrecision)
}

func round(value, precision float64) float64 {
	return math.Round(value/precision) * precision
}

// Only used for debug logging to understand what RainTree is doing under the hood
func (n *rainTreeNetwork) debugMsgTargetString(len, idx int) string {
	s := strings.Builder{}
	s.WriteString("[")
	serviceUrl := n.addrBookMap[n.addrList[0]].ServiceUrl
	if n.addrList[0] == n.selfAddr.String() {
		s.WriteString(fmt.Sprintf(" (%s) ", serviceUrl))
	} else {
		s.WriteString(fmt.Sprintf("(self) %s ", serviceUrl))
	}

	for i := 1; i < len; i++ {
		serviceUrl := n.addrBookMap[n.addrList[i]].ServiceUrl
		if i == idx {
			s.WriteString(fmt.Sprintf(" **%s** ", serviceUrl))
		} else {
			s.WriteString(fmt.Sprintf(" %s ", serviceUrl))
		}
	}
	s.WriteString("]")
	return s.String()
}
