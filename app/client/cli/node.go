package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	nodeCmd := NewNodeCommand()
	rootCmd.AddCommand(nodeCmd)
}

func NewNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Node",
		Short:   "Commands related to node management and operations",
		Aliases: []string{"node", "n"},
	}

	cmd.AddCommand(nodeSaveCommands()...)

	return cmd
}

func nodeSaveCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use: "Save",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("not impl")
			},
			Short: "save a backup of world state",
		},
	}
	return cmds
}
