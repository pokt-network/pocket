//go:build debug

package p2p

import (
	"fmt"
	"os"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/modules"
)

var peerListTableHeader = []string{"Peer ID", "Pokt Address", "ServiceURL"}

// PrintPeerList retrieves the correct peer list using the peerstore provider
// on the bus and then passes this list to printPeerListTable to print the
// list of peers to os.Stdout as a table
func PrintPeerList(bus modules.Bus, routerType RouterType) error {
	var (
		peers           types.PeerList
		routerPlurality = ""
	)

	// TECHDEBT(#811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule.
	pstoreProvider, err := peerstore_provider.GetPeerstoreProvider(bus)
	if err != nil {
		return err
	}

	switch routerType {
	case StakedRouterType:
		// TODO_IN_THIS_COMMIT: what about unstaked peers actors?
		// if !isStaked ...
		pstore, err := pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
		if err != nil {
			return fmt.Errorf("error getting staked peerstore: %v", err)
		}

		peers = pstore.GetPeerList()
	case UnstakedRouterType:
		pstore, err := pstoreProvider.GetUnstakedPeerstore()
		if err != nil {
			return fmt.Errorf("error getting unstaked peerstore: %v", err)
		}

		peers = pstore.GetPeerList()
	case AllRouterTypes:
		routerPlurality = "s"

		// TODO_IN_THIS_COMMIT: what about unstaked peers actors?
		// if !isStaked ...
		stakedPStore, err := pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
		if err != nil {
			return fmt.Errorf("error getting staked peerstore: %v", err)
		}
		unstakedPStore, err := pstoreProvider.GetUnstakedPeerstore()
		if err != nil {
			return fmt.Errorf("error getting unstaked peerstore: %v", err)
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
		return fmt.Errorf("error printing self address: %w", err)
	}

	// NB: Intentionally printing with `fmt` instead of the logger to match
	// `utils.printPeerListTable` which does not use the logger due to
	// incompatibilities with the tabwriter.
	// (This doesn't seem to work as expected; i.e. not printing at all in tilt.)
	if _, err := fmt.Fprintf(
		os.Stdout,
		"%s router peerstore%s:\n",
		routerType,
		routerPlurality,
	); err != nil {
		return fmt.Errorf("error printing to stdout: %w", err)
	}

	if err := printPeerListTable(peers); err != nil {
		return fmt.Errorf("error printing peer list: %w", err)
	}
	return nil
}

// printPeerListTable prints a table of the passed peers to stdout. Header row is defined
// by `peerListTableHeader`. Row printing behavior is defined by `peerListRowConsumerFactory`.
func printPeerListTable(peers types.PeerList) error {
	return utils.PrintTable(peerListTableHeader, peerListRowConsumerFactory(peers))
}

func peerListRowConsumerFactory(peers types.PeerList) utils.RowConsumer {
	return func(provideRow utils.RowProvider) error {
		for _, peer := range peers {
			libp2pAddrInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
			if err != nil {
				return fmt.Errorf("error converting peer to libp2p addr info: %w", err)
			}

			err = provideRow(
				libp2pAddrInfo.ID.String(),
				peer.GetAddress().String(),
				peer.GetServiceURL(),
			)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
