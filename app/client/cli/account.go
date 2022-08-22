package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	privateKeyFilePath string
)

func NewAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Account",
		Short:   "Account specific commands",
		Aliases: []string{"account"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(accountCommands()...)

	return cmd
}

func accountCommands() (cmds []*cobra.Command) {

	sendCmd := &cobra.Command{
		Use:     "Send <from> <to> <amount>",
		Short:   "Send <from> <to> <amount>",
		Long:    "Sends <amount> to address <to> from address <from>",
		Aliases: []string{"send"},
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(deblasis): parse and use privateKeyFilePath
			// TODO(deblasis): implement RPC client, route and handler
			fmt.Printf("sending %s from %s to %s\n", args[2], args[0], args[1])
			return nil
		},
	}

	sendCmd.PersistentFlags().StringVar(&privateKeyFilePath, "path_to_private_key_file", "./pk.json", "Path to private key to use when signing")

	cmds = append(cmds, sendCmd)

	return cmds
}
