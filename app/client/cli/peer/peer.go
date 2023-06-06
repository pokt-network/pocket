package peer

import (
	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/app/client/cli/helpers"
)

var (
	allFlag,
	stakedFlag,
	unstakedFlag bool

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
}
