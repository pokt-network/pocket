package peer

import (
	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/app/client/cli/helpers"
)

var allFlag,
	stakedFlag,
	unstakedFlag,
	localFlag bool

func NewPeerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "Peer",
		Short:             "Manage peers",
		Aliases:           []string{"peer"},
		PersistentPreRunE: helpers.P2PDependenciesPreRunE,
	}

	cmd.PersistentFlags().
		BoolVarP(
			&allFlag,
			"all", "a",
			false,
			"operations apply to both staked & unstaked router peerstores (default)",
		)
	cmd.PersistentFlags().
		BoolVarP(
			&stakedFlag,
			"staked", "s",
			false,
			"operations only apply to staked router peerstore (i.e. raintree)",
		)
	cmd.PersistentFlags().
		BoolVarP(
			&unstakedFlag,
			"unstaked", "u",
			false,
			"operations only apply to unstaked (including staked as a subset) router peerstore (i.e. gossipsub)",
		)
	cmd.PersistentFlags().
		BoolVarP(
			&localFlag,
			"local", "l",
			false,
			"operations apply to the local (CLI binary's) P2P module instead of being broadcast",
		)

	// Add subcommands
	cmd.AddCommand(NewListCommand())

	return cmd
}
