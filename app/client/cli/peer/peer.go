package peer

import (
	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/app/client/cli/helpers"
)

var (
	allFlag,
	stakedFlag,
	unstakedFlag,
	localFlag bool

	PeerCmd = &cobra.Command{
		Use:               "peer",
		Short:             "Manage peers",
		PersistentPreRunE: helpers.P2PDependenciesPreRunE,
	}
)

func init() {
	PeerCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "operations apply to both staked & unstaked router peerstores")
	PeerCmd.PersistentFlags().BoolVarP(&stakedFlag, "staked", "s", false, "operations only apply to staked router peerstore (i.e. raintree)")
	PeerCmd.PersistentFlags().BoolVarP(&unstakedFlag, "unstaked", "u", false, "operations only apply to unstaked router peerstore (i.e. gossipsub)")
	PeerCmd.PersistentFlags().BoolVarP(&localFlag, "local", "l", false, "operations apply to the local (CLI binary's) P2P module rather than being sent to the --remote_cli_url")
}
