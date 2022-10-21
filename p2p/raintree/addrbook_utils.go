package raintree

import (
	"log"
	"math"
	"strings"
)

// Refer to the P2P specification for a formal description and proof of how the constants are selected
const (
	firstMsgTargetPercentage  = float64(1) / float64(3)
	secondMsgTargetPercentage = float64(2) / float64(3)
	shrinkagePercentage       = float64(2) / float64(3)
	maxLevelsLogBase          = float64(3)
	floatPrecision            = float64(0.0000001)
)

// TODO(team): Need to integrate with persistence layer so we are storing this on a per height basis.
// We can easily hit an issue where we are propagating a message from an older height (e.g. before
// the addr book was updated), but we're using `maxNumLevels` associated with the number of
// validators at the current height.
func (n *rainTreeNetwork) getAddrBookLength(level uint32, _height uint64) int {
	peersManagerStateView := n.peersManager.getNetworkView()
	shrinkageCoefficient := math.Pow(shrinkagePercentage, float64(peersManagerStateView.maxNumLevels-level))
	return int(float64(len(peersManagerStateView.addrList)) * (shrinkageCoefficient))
}

// getTargetsAtLevel returns the targets for a given level
func (n *rainTreeNetwork) getTargetsAtLevel(level uint32) []target {
	height := n.GetBus().GetConsensusModule().CurrentHeight()
	addrBookLenghtAtHeight := n.getAddrBookLength(level, height)
	firstTarget := n.getTarget(firstMsgTargetPercentage, addrBookLenghtAtHeight, level)
	secondTarget := n.getTarget(secondMsgTargetPercentage, addrBookLenghtAtHeight, level)

func (n *rainTreeNetwork) getFirstTargetAddr(level int32) cryptoPocket.Address {
	return n.getTarget(level, firstMsgTargetPercentage)
}

func (n *rainTreeNetwork) getSecondTargetAddr(level int32) cryptoPocket.Address {
	return n.getTarget(level, secondMsgTargetPercentage)
}

func (n *rainTreeNetwork) getTarget(level int32, targetPercentage float64) cryptoPocket.Address {
	// OPTIMIZE(olshansky): We are computing this twice for each message, but it's not that expensive.
	l := n.getAddrBookLengthAtHeight(level)

	peersManagerStateView := n.peersManager.getNetworkView()

	target := target{
		serviceUrl:             peersManagerStateView.addrBookMap[peersManagerStateView.addrList[i]].ServiceUrl,
		percentage:             targetPercentage,
		level:                  level,
		addrBookLengthAtHeight: addrBookLen,
		index:                  i,
	}

	// If the target is 0, it is a reference to self, which is a `Demote` in RainTree terms.
	// This is handled separately.
	if i == 0 {
		target.isSelf = true
		return target
	}

	addrStr := peersManagerStateView.addrList[i]
	if addr, ok := peersManagerStateView.addrBookMap[addrStr]; ok {
		target.address = addr.Address
		return target
	}
	log.Printf("[DEBUG] addrStr %s not found in addrBookMap", addrStr)
	return target
}

// Only used for debug logging to understand what RainTree is doing under the hood
func (n *rainTreeNetwork) debugMsgTargetString(target1, target2 target) string {
	s := strings.Builder{}
	s.WriteString(target1.DebugString(n))
	s.WriteString(" --|-- ")
	s.WriteString(target2.DebugString(n))
	return s.String()
}
