package peer

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/app/client/cli/helpers"
	"github.com/pokt-network/pocket/p2p/debug"
	"github.com/pokt-network/pocket/shared/messaging"
)

var (
	connectionsCmd = &cobra.Command{
		Use:   "connections",
		Short: "Print open peer connections",
		RunE:  connectionsRunE,
	}
)

func init() {
	PeerCmd.AddCommand(connectionsCmd)
}

func connectionsRunE(cmd *cobra.Command, _ []string) error {
	var routerType debug.RouterType

	bus, err := helpers.GetBusFromCmd(cmd)
	if err != nil {
		return err
	}

	switch {
	case stakedFlag:
		if unstakedFlag || allFlag {
			return ErrRouterType
		}
		routerType = debug.StakedRouterType
	case unstakedFlag:
		if stakedFlag || allFlag {
			return ErrRouterType
		}
		routerType = debug.UnstakedRouterType
	// even if `allFlag` is false, we still want to print all connections
	default:
		if stakedFlag || unstakedFlag {
			return ErrRouterType
		}
		routerType = debug.AllRouterTypes
	}

	debugMsg := &messaging.DebugMessage{
		Action: messaging.DebugMessageAction_DEBUG_P2P_PEER_CONNECTIONS,
		Type:   messaging.DebugMessageRoutingType_DEBUG_MESSAGE_TYPE_BROADCAST,
		Message: &anypb.Any{
			Value: []byte(routerType),
		},
	}
	debugMsgAny, err := anypb.New(debugMsg)
	if err != nil {
		return fmt.Errorf("creating anypb from debug message: %w", err)
	}

	if localFlag {
		if err := debug.PrintPeerConnections(bus, routerType); err != nil {
			return fmt.Errorf("printing peer list: %w", err)
		}
		return nil
	}

	// TECHDEBT(#810, #811): will need to wait for DHT bootstrapping to complete before
	// p2p broadcast can be used with to reach unstaked actors.
	// CONSIDERATION: add the peer commands to the interactive CLI as the P2P module
	// instance could persist between commands. Other interactive CLI commands which
	// rely on unstaked actor router broadcast are working as expected.

	// TECHDEBT(#810, #811): use broadcast instead to reach all peers.
	return sendToStakedPeers(cmd, debugMsgAny)
}
