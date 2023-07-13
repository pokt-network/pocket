package debug

import (
	"fmt"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
	"os"
)

func PrintPeerList(bus modules.Bus, routerType RouterType) error {
	var (
		peers           types.PeerList
		pstorePlurality = ""
	)

	// TECHDEBT(#810, #811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule.
	pstoreProviderModule, err := bus.GetModulesRegistry().
		GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	if err != nil {
		return fmt.Errorf("getting peerstore provider: %w", err)
	}
	pstoreProvider, ok := pstoreProviderModule.(peerstore_provider.PeerstoreProvider)
	if !ok {
		return fmt.Errorf("unknown peerstore provider type: %T", pstoreProviderModule)
	}
	//--

	switch routerType {
	case StakedRouterType:
		// TODO_THIS_COMMIT: what about unstaked peers actors?
		// if !isStaked ...
		pstore, err := pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
		if err != nil {
			return fmt.Errorf("getting unstaked peerstore: %v", err)
		}

		peers = pstore.GetPeerList()
	case UnstakedRouterType:
		pstore, err := pstoreProvider.GetUnstakedPeerstore()
		if err != nil {
			return fmt.Errorf("getting unstaked peerstore: %v", err)
		}

		peers = pstore.GetPeerList()
	case AllRouterTypes:
		pstorePlurality = "s"

		// TODO_THIS_COMMIT: what about unstaked peers actors?
		// if !isStaked ...
		stakedPStore, err := pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
		if err != nil {
			return fmt.Errorf("getting unstaked peerstore: %v", err)
		}
		unstakedPStore, err := pstoreProvider.GetUnstakedPeerstore()
		if err != nil {
			return fmt.Errorf("getting unstaked peerstore: %v", err)
		}

		unstakedPeers := unstakedPStore.GetPeerList()
		stakedPeers := stakedPStore.GetPeerList()
		additionalPeers, _ := unstakedPeers.Delta(stakedPeers)

		// NB: there should never be any "additional" peers. This would represent
		// a staked actor who is not participating in background gossip for some
		// reason. It's possible that a staked actor node which has restarted
		// recently and hasn't yet completed background router bootstrapping may
		// result in peers experiencing this state.
		if len(additionalPeers) == 0 {
			return PrintPeerListTable(unstakedPeers)
		}

		allPeers := append(types.PeerList{}, unstakedPeers...)
		allPeers = append(allPeers, additionalPeers...)
		peers = allPeers
	default:
		return fmt.Errorf("unsupported router type: %s", routerType)
	}

	if err := LogSelfAddress(bus); err != nil {
		return fmt.Errorf("printing self address: %w", err)
	}

	// NB: Intentionally printing with `fmt` instead of the logger to match
	// `utils.PrintPeerListTable` which does not use the logger due to
	// incompatibilities with the tabwriter.
	// (This doesn't seem to work as expected; i.e. not printing at all in tilt.)
	if _, err := fmt.Fprintf(
		os.Stdout,
		"%s router peerstore%s:\n",
		routerType,
		pstorePlurality,
	); err != nil {
		return fmt.Errorf("printing to stdout: %w", err)
	}

	if err := PrintPeerListTable(peers); err != nil {
		return fmt.Errorf("printing peer list: %w", err)
	}
	return nil
}

func getPeerstoreProvider() (peerstore_provider.PeerstoreProvider, error) {
	return nil, nil
}