package raintree

import (
	"math"
	"strconv"
)

// Refer to the P2P specification for a formal description and proof of how the constants are selected
const (
	firstMsgTargetPercentage  = float64(1) / float64(3)
	secondMsgTargetPercentage = float64(2) / float64(3)
	shrinkagePercentage       = float64(2) / float64(3)
	maxLevelsLogBase          = float64(3)
	floatPrecision            = float64(0.0000001)
)

func (rtr *rainTreeRouter) getPeerstoreSize(level uint32, height uint64) int {
	peersView, maxNumLevels := rtr.peersManager.getPeersViewWithLevels()

	// if we are propagating a message from a previous height, we need to instantiate an ephemeral rainTreePeersManager (without add/remove)
	if height < rtr.currentHeightProvider.CurrentHeight() {
		peersMgr, err := newPeersManagerWithPeerstoreProvider(rtr.selfAddr, rtr.pstoreProvider, height)
		if err != nil {
			rtr.logger.Fatal().Err(err).Msg("Error initializing rainTreeRouter rainTreePeersManager")
		}
		peersView, maxNumLevels = peersMgr.getPeersViewWithLevels()
	}

	shrinkageCoefficient := math.Pow(shrinkagePercentage, float64(maxNumLevels-level))
	return int(float64(len(peersView.GetAddrs())) * (shrinkageCoefficient))
}

// getTargetsAtLevel returns the targets for a given level
func (rtr *rainTreeRouter) getTargetsAtLevel(level uint32) []target {
	height := rtr.currentHeightProvider.CurrentHeight()
	pstoreSizeAtHeight := rtr.getPeerstoreSize(level, height)
	firstTarget := rtr.getTarget(firstMsgTargetPercentage, pstoreSizeAtHeight, level)
	secondTarget := rtr.getTarget(secondMsgTargetPercentage, pstoreSizeAtHeight, level)

	rtr.logger.Debug().Fields(
		map[string]any{
			"firstTarget":  firstTarget.serviceURL,
			"secondTarget": secondTarget.serviceURL,
			"height":       height,
			"level":        strconv.Itoa(int(level)), // TECHDEBT(#): Figure out why we need a conversion here
			"pstoreSize":   pstoreSizeAtHeight,
		},
	).Msg("Targets at height")

	return []target{firstTarget, secondTarget}
}

func (rtr *rainTreeRouter) getTarget(targetPercentage float64, pstoreSize int, level uint32) target {
	i := int(targetPercentage * float64(pstoreSize))
	peersView := rtr.peersManager.GetPeersView()
	serviceURL := peersView.GetPeers()[i].GetServiceURL()

	target := target{
		serviceURL:            serviceURL,
		percentage:            targetPercentage,
		level:                 level,
		peerstoreSizeAtHeight: pstoreSize,
		index:                 i,
	}

	// If the target is 0, it is a reference to self, which is a `Demote` in RainTree terms.
	// This is handled separately.
	if i == 0 {
		target.isSelf = true
		return target
	}

	addrStr := peersView.GetAddrs()[i]
	if addr := rtr.GetPeerstore().GetPeerFromString(addrStr); addr != nil {
		target.address = addr.GetAddress()
		return target
	}

	rtr.logger.Debug().Str("address", addrStr).Msg("address not found in Peerstore")

	return target
}
