package cli

import (
	"fmt"
	"path/filepath"

	"github.com/pokt-network/pocket/app/client/keybase"
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
			Use:     "List",
			Short:   "List all keys",
			Long:    "List all the public hex addresses for the keys in the keybase",
			Aliases: []string{"list"},
			Args:    cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				keybaseDir, err := filepath.Abs(dataDir + "/keys")
				if err != nil {
					return err
				}
				kb, err := keybase.InitialiseKeybase(keybaseDir)
				if err != nil {
					return err
				}
				addresses, _, err := kb.GetAll()
				if err != nil {
					return err
				}
				if err := kb.Stop(); err != nil {
					return err
				}
				fmt.Println("Public Key Addresses:")
				for _, addr := range addresses {
					fmt.Println(addr)
				}

				return nil
			},
		},
	}
	for _, cmd := range cmds {
		cmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	}
	return cmds
}
