package cli

import (
	"fmt"
	"strings"

	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

type actorCmdDef struct {
	Name      string
	ActorType types.ActorType
}

func NewActorCommands() (cmds []*cobra.Command) {

	actorCmdDefs := []actorCmdDef{
		{"Application", types.ActorType_App},
		{"Node", types.ActorType_Node},
		{"Fisherman", types.ActorType_Fish},
		{"Validator", types.ActorType_Val},
	}

	for _, cmdDef := range actorCmdDefs {

		cmd := &cobra.Command{
			Use:     cmdDef.Name,
			Short:   fmt.Sprintf("%s actor specific commands", cmdDef.Name),
			Aliases: []string{strings.ToLower(cmdDef.Name), cmdDef.ActorType.GetActorName()},
			Args:    cobra.ExactArgs(0),
		}
		cmd.AddCommand(newActorCommands(cmdDef)...)
		cmds = append(cmds, cmd)
	}

	return cmds
}

func newActorCommands(cmdDef actorCmdDef) (cmds []*cobra.Command) {

	stakeCmd := &cobra.Command{
		Use:   "Stake <from> <amount>",
		Short: "Stake <from> <amount>",
		Long:  fmt.Sprintf(`Stakes <amount> for the %s actot with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendRPC("stake", cmdDef.ActorType, args)(cmd, args)
		},
	}
	cmds = append(cmds, stakeCmd)

	editStakeCmd := &cobra.Command{
		Use:   "EditStake <from> <amount>",
		Short: "EditStake <from> <amount>",
		Long:  fmt.Sprintf(`Stakes a new <amount> for the %s actor with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendRPC("editstake", cmdDef.ActorType, args)(cmd, args)
		},
	}
	cmds = append(cmds, editStakeCmd)

	unstakeCmd := &cobra.Command{
		Use:   "Unstake <from>",
		Short: "Unstake <from>",
		Long:  fmt.Sprintf(`Unstakes the prevously staked tokens for the %s actor with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendRPC("unstake", cmdDef.ActorType, args)(cmd, args)
		},
	}
	cmds = append(cmds, unstakeCmd)

	return cmds
}

func sendRPC(rpcType string, actorType types.ActorType, args []string) func(cmd *cobra.Command, args []string) error {
	// TODO(deblasis): refactor this placeholder
	// TODO(deblasis): implement RPC client, route and handler
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}
