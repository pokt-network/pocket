package cli

import "github.com/spf13/cobra"

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

	return cmd
}
