package peerstore_provider

import (
	"fmt"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// unstakedPeerstoreProvider is an interface which the p2p module supports in
// order to allow access to the unstaked-actor-router's peerstore
//
// TECHDEBT(#xxx): will become unnecessary after `modules.P2PModule#GetUnstakedPeerstore` is added.`
// CONSIDERATION: split `PeerstoreProvider` into `StakedPeerstoreProvider` and `UnstakedPeerstoreProvider`.
// (see: https://github.com/pokt-network/pocket/pull/804#issuecomment-1576531916)
type unstakedPeerstoreProvider interface {
	GetUnstakedPeerstore() (typesP2P.Peerstore, error)
}

func GetUnstakedPeerstore(bus modules.Bus) (typesP2P.Peerstore, error) {
	p2pModule := bus.GetP2PModule()
	if p2pModule == nil {
		return nil, fmt.Errorf("p2p module is not registered to bus and is required")
	}

	unstakedPSP, ok := p2pModule.(unstakedPeerstoreProvider)
	if !ok {
		return nil, fmt.Errorf("p2p module does not implement unstakedPeerstoreProvider")
	}
	return unstakedPSP.GetUnstakedPeerstore()
}
