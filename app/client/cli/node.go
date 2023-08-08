package cli

import (
	"fmt"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/rpc"
	"github.com/spf13/cobra"
)

func init() {
	nodeCmd := NewNodeCommand()
	rootCmd.AddCommand(nodeCmd)
}

var (
	dir string
)

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
			Use:     "Save",
			Short:   "save a backup of node databases in the provided directory",
			Example: "node save --dir /dir/path/here/",
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(flags.RemoteCLIURL)
				if err != nil {
					return err
				}
				resp, err := client.PostV1NodeBackup(cmd.Context(), rpc.NodeBackup{
					Dir: &dir,
				})
				if err != nil {
					return err
				}
				var dest []byte
				_, err = resp.Body.Read(dest)
				if err != nil {
					return err
				}
				fmt.Printf("%s", dest)
				return nil
			},
		},
	}
	return cmds
}
