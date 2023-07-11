package raintree

import (
	"fmt"
	"math"
	"strconv"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
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

	// TECHDEBT(#810, 811): use `bus.GetPeerstoreProvider()` instead once available.
	pstoreProvider, err := rtr.getPeerstoreProvider()
	if err != nil {
		// Should never happen; enforced by a `rtr.getPeerstoreProvider()` call
		// & error handling in `rtr.broadcastAtLevel()`.
		panic(fmt.Sprintf("Error retrieving peerstore provider: %s", err.Error()))
	}

	// if we are propagating a message from a previous height, we need to instantiate an ephemeral rainTreePeersManager (without add/remove)
	if height < rtr.GetBus().GetCurrentHeightProvider().CurrentHeight() {
		peersMgr, err := newPeersManagerWithPeerstoreProvider(rtr.selfAddr, pstoreProvider, height)
		if err != nil {
			rtr.logger.Fatal().Err(err).Msg("Error initializing rainTreeRouter rainTreePeersManager")
		}
		peersView, maxNumLevels = peersMgr.getPeersViewWithLevels()
	}

	shrinkageCoefficient := math.Pow(shrinkagePercentage, float64(maxNumLevels-level))
	return int(float64(len(peersView.GetAddrs())) * (shrinkageCoefficient))
}

// TECHDEBT(#810, 811): replace with `bus.GetPeerstoreProvider()` once available.
func (rtr *rainTreeRouter) getPeerstoreProvider() (peerstore_provider.PeerstoreProvider, error) {
	pstoreProviderModule, err := rtr.GetBus().GetModulesRegistry().
		GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	if err != nil {
		return nil, err
	}

	pstoreProvider, ok := pstoreProviderModule.(peerstore_provider.PeerstoreProvider)
	if !ok {
		return nil, fmt.Errorf("unexpected peerstore provider module type: %T", pstoreProviderModule)
	}
	return pstoreProvider, nil
}

// getTargetsAtLevel returns the targets for a given level
func (rtr *rainTreeRouter) getTargetsAtLevel(level uint32) []target {
	height := rtr.GetBus().GetCurrentHeightProvider().CurrentHeight()
	pstoreSizeAtHeight := rtr.getPeerstoreSize(level, height)
	firstTarget := rtr.getTarget(firstMsgTargetPercentage, pstoreSizeAtHeight, level)
	secondTarget := rtr.getTarget(secondMsgTargetPercentage, pstoreSizeAtHeight, level)

	rtr.logger.Debug().Fields(
		map[string]any{
			"firstTarget":  firstTarget.serviceURL,
			"secondTarget": secondTarget.serviceURL,
			"height":       height,
			"level":        strconv.Itoa(int(level)), // HACK(#783): Figure out why we need a conversion here
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
