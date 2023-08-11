//go:build debug

package peer

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/app/client/cli/helpers"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
)

var ErrRouterType = fmt.Errorf("must specify one (or none) of --staked, --unstaked, --libp2p_host, or --all")

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "List",
		Short:   "List the known peers",
		Long:    "Prints a table of the Peer ID, Pokt Address and Service URL of the known peers",
		Aliases: []string{"list", "ls"},
		RunE:    listRunE,
	}
}

func listRunE(cmd *cobra.Command, _ []string) error {
	// TODO_THIS_COMMIT: comment; explain
	time.Sleep(500 * time.Millisecond)

	var routerType p2p.RouterType

	bus, err := helpers.GetBusFromCmd(cmd)
	if err != nil {
		return err
	}

	switch {
	// --staked
	case stakedFlag && !unstakedFlag && !allFlag && !libp2pHostFlag:
		routerType = p2p.StakedRouterType
	// --unstaked
	case unstakedFlag && !stakedFlag && !allFlag && !libp2pHostFlag:
		routerType = p2p.UnstakedRouterType
	// --libp2p_host
	case libp2pHostFlag && !unstakedFlag && !stakedFlag && !allFlag:
		routerType = p2p.Libp2pHost
	// --all (--staked | --unstaked)
	// --libp2p_host (--staked | --unstaked)
	case (stakedFlag || unstakedFlag) && (allFlag || libp2pHostFlag):
		return ErrRouterType
	// even if `allFlag` is false, we still want to print all
	// --all
	default:
		routerType = p2p.AllRouterTypes
	}

	logger.Global.Debug().Msgf("DEBUG: router type: %s", routerType)

	debugMsg := &messaging.DebugMessage{
		Action: messaging.DebugMessageAction_DEBUG_P2P_PRINT_PEER_LIST,
		Type:   messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
		Message: &anypb.Any{
			Value: []byte(routerType),
		},
	}
	debugMsgAny, err := anypb.New(debugMsg)
	if err != nil {
		return fmt.Errorf("error creating anypb from debug message: %w", err)
	}

	if localFlag {
		if err := p2p.PrintPeerList(bus, routerType); err != nil {
			return fmt.Errorf("error printing peer list: %w", err)
		}
		// TODO_THIS_COMMIT: uncomment; intended to temporarily print the local peer list at the same time as the remote.
		//return nil
	}

	remotePeer, err := remotePeerFromRemoteCLIURLFlag(bus)
	if err != nil {
		return err
	}

	// If broadcast flag is not set, send debug message directly to the peer
	// corresponding to the value of the --remote_cli_url flag.
	if !broadcastFlag {
		if err := bus.GetP2PModule().Send(remotePeer.GetAddress(), debugMsgAny); err != nil {
			return err
		}
		return nil
	}

	// TECHDEBT(#811): will need to wait for DHT bootstrapping to complete before
	// p2p broadcast can be used with to reach unstaked actors.
	// CONSIDERATION: add the peer commands to the interactive CLI as the P2P module
	// instance could persist between commands. Other interactive CLI commands which
	// rely on unstaked actor router broadcast are working as expected.

	// TECHDEBT(#811): use broadcast instead to reach all peers.
	err = sendToStakedPeers(cmd, debugMsgAny)

	time.Sleep(500 * time.Millisecond)

	return err
}

func sendToStakedPeers(cmd *cobra.Command, debugMsgAny *anypb.Any) error {
	bus, err := helpers.GetBusFromCmd(cmd)
	if err != nil {
		return err
	}

	pstore, err := helpers.FetchPeerstore(cmd)
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("unable to retrieve the pstore")
	}

	if pstore.Size() == 0 {
		logger.Global.Fatal().Msg("no validators found")
	}

	// TODO_THIS_COMMIT: choose & cleanup
	if err := bus.GetP2PModule().Broadcast(debugMsgAny); err != nil {
		logger.Global.Error().Err(err).Msg("failed to send debug message")
	}

	//for _, peer := range pstore.GetPeerList() {
	//	if err := bus.GetP2PModule().Send(peer.GetAddress(), debugMsgAny); err != nil {
	//		logger.Global.Error().Err(err).Msg("failed to send debug message")
	//	}
	//}
	return nil
}

// remotePeerFromRemoteCLIURLFlag returns the pokt address that corresponds to
// the peer specified by the value of the --remote_cli_url flag.
func remotePeerFromRemoteCLIURLFlag(bus modules.Bus) (typesP2P.Peer, error) {
	pstoreProviderModule, err := bus.GetModulesRegistry().GetModule(peerstore_provider.PeerstoreProviderSubmoduleName)
	pstoreProvider, ok := pstoreProviderModule.(peerstore_provider.PeerstoreProvider)
	if !ok {
		return nil, fmt.Errorf("error unexpected peerstore provider module type: %T", pstoreProviderModule)
	}

	pstore, err := pstoreProvider.GetStakedPeerstoreAtCurrentHeight()
	if err != nil {
		return nil, err
	}

	var remotePeer typesP2P.Peer
	for _, peer := range pstore.GetPeerList() {
		if flags.RemoteCLIURL == peer.GetServiceURL() {
			remotePeer = peer
			break
		}
	}
	return remotePeer, nil
}
