package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewGovernanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Governance",
		Short:   "Governance specific commands",
		Aliases: []string{"gov"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(govCommands()...)

	return cmd
}

func govCommands() (cmds []*cobra.Command) {

	cmds = append(cmds, &cobra.Command{
		Use:   "ChangeParameter <owner> <key> <value>",
		Short: "ChangeParameter <owner> <key> <value>",
		// DISCUSS(deblasis): do we need some sort of validation on the backend?
		Long:    "Changes the Governance parameter with <key> owned by <owner> to <value>",
		Aliases: []string{},
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(deblasis): implement RPC client, route and handler
			fmt.Printf("changing parameter %s owned by %s to %s\n", args[1], args[0], args[2])
			return nil
		},
	})

	return cmds
}
