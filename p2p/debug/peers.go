package debug

import (
	"fmt"
	"os"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/modules"
)

type RouterType string

const (
	StakedRouterType   RouterType = "staked"
	UnstakedRouterType RouterType = "unstaked"
	AllRouterTypes     RouterType = "all"
)

var peerListTableHeader = []string{"Peer ID", "Pokt Address", "ServiceURL"}

func LogSelfAddress(bus modules.Bus) error {
	p2pModule := bus.GetP2PModule()
	if p2pModule == nil {
		return fmt.Errorf("no p2p module found on the bus")
	}

	selfAddr, err := p2pModule.GetAddress()
	if err != nil {
		return fmt.Errorf("getting self address: %w", err)
	}

	_, err = fmt.Fprintf(os.Stdout, "self address: %s", selfAddr.String())
	return err
}

// PrintPeerListTable prints a table of the passed peers to stdout. Header row is defined
// by `peerListTableHeader`. Row printing behavior is defined by `peerListRowConsumerFactory`.
func PrintPeerListTable(peers types.PeerList) error {
	return utils.PrintTable(peerListTableHeader, peerListRowConsumerFactory(peers))
}

func peerListRowConsumerFactory(peers types.PeerList) utils.RowConsumer {
	return func(provideRow utils.RowProvider) error {
		for _, peer := range peers {
			libp2pAddrInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
			if err != nil {
				return fmt.Errorf("converting peer to libp2p addr info: %w", err)
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
