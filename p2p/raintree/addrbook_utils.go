package raintree

import (
	"fmt"
	"log"
	"math"
	"strings"

	typesP2P "github.com/pokt-network/pocket/p2p/types"

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
func (n *rainTreeNetwork) processAddrBookUpdates() error {
	var err error
	n.maxNumLevels = n.getMaxAddrBookLevels()
	n.addrList, n.addrBookMap, err = n.addrBook.ToListAndMap(n.selfAddr.String()) // TODO (Team) stick to convention, string or bytes
	if err != nil {
		return err
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

func (n *rainTreeNetwork) getFirstTargetAddr(level int32) cryptoPocket.Address {
	return n.getTarget(level, firstMsgTargetPercentage)
}

func (n *rainTreeNetwork) getSecondTargetAddr(level int32) cryptoPocket.Address {
	return n.getTarget(level, secondMsgTargetPercentage)
}

func (n *rainTreeNetwork) getTarget(level int32, targetPercentage float64) cryptoPocket.Address {
	// OPTIMIZE(olshansky): We are computing this twice for each message, but it's not that expensive.
	l := n.getAddrBookLengthAtHeight(level)

	i := int(targetPercentage * float64(l))

	// If the target is 0, it is a reference to self, which is a `Demote` in RainTree terms.
	// This is handled separately.
	if i == 0 {
		return nil
	}

	addrStr := n.addrList[i]
	if addr, ok := n.addrBookMap[addrStr]; ok {
		// IMPROVE(olshansky): Consolidate so the debug print contains all (i.e. both) targets in one log line
		log.Printf("[DEBUG] Target (%0.2f) at height (%d): %s", targetPercentage, level, n.debugMsgTargetString(l, i))
		return addr.Address
	}
	return nil
}

func (n *rainTreeNetwork) getLeftAndRight() (left cryptoPocket.Address, right cryptoPocket.Address, ok bool) {
	return getLeftAndRight(n.addrList, n.addrBookMap)
}

// TODO (team): should make addrList a type
func getLeftAndRight(addrList []string, addrBookMap typesP2P.AddrBookMap) (left cryptoPocket.Address, right cryptoPocket.Address, ok bool) {
	if len(addrList) < 3 {
		return nil, nil, false
	}
	leftAddress := addrList[len(addrList)-1]
	rightAddress := addrList[1]
	return addrBookMap[leftAddress].Address, addrBookMap[rightAddress].Address, true
}

// TODO(team): Need to integrate with persistence layer so we are storing this on a per height basis.
// We can easily hit an issue where we are propagating a message from an older height (e.g. before
// the addr book was updated), but we're using `maxNumLevels` associated with the number of
// validators at the current height.
func (n *rainTreeNetwork) getAddrBookLengthAtHeight(level int32) int {
	shrinkageCoefficient := math.Pow(shrinkagePercentage, float64(n.maxNumLevels-level))
	return int(float64(len(n.addrList)) * (shrinkageCoefficient))
}

func (n *rainTreeNetwork) getMaxAddrBookLevels() int32 {
	return getMaxAddrBookLevels(n.addrBook)
}

func getMaxAddrBookLevels(addrBook typesP2P.AddrBook) int32 {
	addrBookSize := float64(len(addrBook))
	return int32(math.Ceil(logBase(addrBookSize)))
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
