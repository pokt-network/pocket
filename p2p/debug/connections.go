package debug

import (
	"fmt"
	"os"
	"strconv"

	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/modules"
)

var printConnectionsHeader = []string{"Peer ID", "Multiaddr", "Opened", "Direction", "NumStreams"}

func PrintPeerConnections(bus modules.Bus, routerType RouterType) error {
	var (
		connections     []libp2pNetwork.Conn
		routerPlurality = ""
	)

	if routerType == AllRouterTypes {
		routerPlurality = "s"
	}

	connections, err := getFilteredConnections(bus, routerType)
	if err != nil {
		return fmt.Errorf("getting connecions: %w", err)
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
		routerPlurality,
	); err != nil {
		return fmt.Errorf("printing to stdout: %w", err)
	}

	if err := PrintConnectionsTable(connections); err != nil {
		return fmt.Errorf("printing peer list: %w", err)
	}
	return nil
}

func PrintConnectionsTable(conns []libp2pNetwork.Conn) error {
	return utils.PrintTable(printConnectionsHeader, peerConnsRowConsumerFactory(conns))
}

func getFilteredConnections(
	bus modules.Bus,
	routerType RouterType,
) ([]libp2pNetwork.Conn, error) {
	var (
		pstore       typesP2P.Peerstore
		idsToInclude map[libp2pPeer.ID]struct{}
		p2pModule    = bus.GetP2PModule()
		connections  = p2pModule.GetConnections()
	)

	// TECHDEBT(#810, #811): use `bus.GetPeerstoreProvider()` after peerstore provider
	// is retrievable as a proper submodule.
	pstoreProviderModule, err := bus.GetModulesRegistry().
		GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	if err != nil {
		return nil, fmt.Errorf("getting peerstore provider: %w", err)
	}
	pstoreProvider, ok := pstoreProviderModule.(peerstore_provider.PeerstoreProvider)
	if !ok {
		return nil, fmt.Errorf("unknown peerstore provider type: %T", pstoreProviderModule)
	}
	//--

	switch routerType {
	case AllRouterTypes:
		// return early; no need to filter
		return connections, nil
	case StakedRouterType:
		pstore, err = pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
		if err != nil {
			return nil, fmt.Errorf("getting staked peerstore: %w", err)
		}
	case UnstakedRouterType:
		pstore, err = pstoreProvider.GetUnstakedPeerstore()
		if err != nil {
			return nil, fmt.Errorf("getting unstaked peerstore: %w", err)
		}
	}

	idsToInclude, err = getPeerIDs(pstore.GetPeerList())
	if err != nil {
		return nil, fmt.Errorf("getting peer IDs: %w", err)
	}

	var filteredConnections []libp2pNetwork.Conn
	for _, conn := range connections {
		if _, ok := idsToInclude[conn.RemotePeer()]; ok {
			filteredConnections = append(filteredConnections, conn)
		}
	}
	return filteredConnections, nil
}

func peerConnsRowConsumerFactory(conns []libp2pNetwork.Conn) utils.RowConsumer {
	return func(provideRow utils.RowProvider) error {
		for _, conn := range conns {
			if err := provideRow(
				conn.RemotePeer().String(),
				conn.RemoteMultiaddr().String(),
				conn.Stat().Opened.String(),
				conn.Stat().Direction.String(),
				strconv.Itoa(conn.Stat().NumStreams),
			); err != nil {
				return err
			}
		}
		return nil
	}
}

func getPeerIDs(peers []typesP2P.Peer) (map[libp2pPeer.ID]struct{}, error) {
	ids := make(map[libp2pPeer.ID]struct{})
	for _, peer := range peers {
		addrInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
		if err != nil {
			return nil, err
		}

		// ID already in set; continue
		if _, ok := ids[addrInfo.ID]; ok {
			continue
		}

		// add ID to set
		ids[addrInfo.ID] = struct{}{}
	}
	return ids, nil
}
