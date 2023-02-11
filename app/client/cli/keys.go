package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	keybaseCmd := NewAccountCommand()
	rootCmd.AddCommand(keybaseCmd)
}

func NewKeybaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Keys",
		Short:   "Keybase specific commands",
		Aliases: []string{"keys"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(keybaseCommands()...)

	return cmd
}

func keybaseCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Send <fromAddr> <to> <amount>",
			Short:   "Send <fromAddr> <to> <amount>",
			Long:    "Sends <amount> to address <to> from address <fromAddr>",
			Aliases: []string{"send"},
			Args:    cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		},
	}
	for _, cmd := range cmds {
		cmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	}
	return cmds
}
