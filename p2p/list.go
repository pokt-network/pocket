//go:build debug

package p2p

import (
	"fmt"
	"os"

	libp2pPeerstore "github.com/libp2p/go-libp2p/core/peerstore"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	peerListTableHeader       = []string{"Peer ID", "Pokt Address", "ServiceURL", "Multiaddr"}
	libp2pPeerListTableHeader = []string{"Peer ID", "Multiaddr"}
)

// PrintPeerList retrieves the correct peer list using the peerstore provider
// on the bus and then passes this list to printPeerListTable to print the
// list of peers to os.Stdout as a table
func PrintPeerList(bus modules.Bus, routerType RouterType) error {
	var (
		peers           typesP2P.PeerList
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
	case Libp2pHost:
		p2pModule := bus.GetP2PModule()
		p2p, ok := p2pModule.(*P2PModule)
		if !ok {
			return fmt.Errorf("unsupported P2P module type: %T", p2pModule)
		}

		_, err = fmt.Fprintf(
			os.Stdout,
			"self peer ID: %s\n",
			p2p.GetLibp2pHost().ID().String(),
		)
		if err != nil {
			return err
		}

		return printLibP2PPeerListTable(p2p.GetLibp2pHost().Peerstore())
	case AllRouterTypes:
		routerPlurality = "s"

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
			peers = unstakedPeers
			break
		}

		allPeers := append(typesP2P.PeerList{}, unstakedPeers...)
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
		return fmt.Errorf("error printing libp2pPeer list: %w", err)
	}
	return nil
}

// printPeerListTable prints a table of the passed peers to stdout. Header row is defined
// by `peerListTableHeader`. Row printing behavior is defined by `peerListRowConsumerFactory`.
func printPeerListTable(peers typesP2P.PeerList) error {
	return utils.PrintTable(peerListTableHeader, peerListRowConsumerFactory(peers))
}

func peerListRowConsumerFactory(peers typesP2P.PeerList) utils.RowConsumer {
	return func(provideRow utils.RowProvider) error {
		for _, peer := range peers {
			libp2pAddrInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
			if err != nil {
				return fmt.Errorf("error converting libp2pPeer to libp2p addr info: %w", err)
			}

			peerMultiaddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.GetServiceURL())
			if err != nil {
				return fmt.Errorf("error converting libp2pPeer service URL to libp2p multiaddr: %w", err)
			}

			err = provideRow(
				libp2pAddrInfo.ID.String(),
				peer.GetAddress().String(),
				peer.GetServiceURL(),
				peerMultiaddr.String(),
			)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// printPeerListTable prints a table of the passed peers to stdout. Header row is defined
// by `peerListTableHeader`. Row printing behavior is defined by `peerListRowConsumerFactory`.
func printLibP2PPeerListTable(pstore libp2pPeerstore.Peerstore) error {
	return utils.PrintTable(libp2pPeerListTableHeader, libp2pPeerListRowConsumerFactory(pstore))
}

func libp2pPeerListRowConsumerFactory(pstore libp2pPeerstore.Peerstore) utils.RowConsumer {
	return func(provideRow utils.RowProvider) error {
		for _, peerID := range pstore.Peers() {
			peerAddrs := pstore.Addrs(peerID)

			peerMultiaddrStr := "empty"
			if len(peerAddrs) > 0 {
				peerMultiaddrStr = peerAddrs[0].String()
			}

			if err := provideRow(
				peerID.String(),
				peerMultiaddrStr,
			); err != nil {
				return err
			}
		}
		return nil
	}
}
