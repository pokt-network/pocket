package raintree

import (
	"log"
	"strings"
)

// Refer to the P2P specification for a formal description and proof of how the constants are selected
const (
	firstMsgTargetPercentage  = float64(1) / float64(3)
	secondMsgTargetPercentage = float64(2) / float64(3)
)

type router struct {
	network *rainTreeNetwork
}

func newRouter(n *rainTreeNetwork) *router {
	return &router{
		network: n,
	}
}

// GetTargetsAtLevel returns the targets for a given level
func (r *router) GetTargetsAtLevel(level uint32) []target {
	addrBookLenghtAtHeight := r.network.getAddrBookLengthAtHeight(level)
	firstTarget := r.getTarget(firstMsgTargetPercentage, addrBookLenghtAtHeight, level)
	secondTarget := r.getTarget(secondMsgTargetPercentage, addrBookLenghtAtHeight, level)

	log.Printf("[DEBUG] Targets at height (%d): %s", level, r.debugMsgTargetString(firstTarget, secondTarget))

	return []target{firstTarget, secondTarget}
}

func (r *router) getTarget(targetPercentage float64, len int, level uint32) target {
	i := int(targetPercentage * float64(len))

	target := target{
		serviceUrl:             r.network.addrBookMap[r.network.addrList[i]].ServiceUrl,
		percentage:             targetPercentage,
		level:                  level,
		addrBookLengthAtHeight: len,
		index:                  i,
	}

	// If the target is 0, it is a reference to self, which is a `Demote` in RainTree terms.
	// This is handled separately.
	if i == 0 {
		target.isSelf = true
		return target
	}

	addrStr := r.network.addrList[i]
	if addr, ok := r.network.addrBookMap[addrStr]; ok {
		target.address = addr.Address
		return target
	}
	log.Printf("[DEBUG] addrStr %s not found in addrBookMap", addrStr)
	return target
}

// Only used for debug logging to understand what RainTree is doing under the hood
func (r *router) debugMsgTargetString(target1, target2 target) string {
	s := strings.Builder{}
	s.WriteString(target1.DebugString(r))
	s.WriteString(" --|-- ")
	s.WriteString(target2.DebugString(r))
	return s.String()
}
