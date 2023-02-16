package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewKeysCommand())
}

func NewKeysCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Keys",
		Short:   "Key specific commands",
		Aliases: []string{"keys"},
		Args:    cobra.ExactArgs(0),
	}

	cmds := keysCommands()
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	cmd.AddCommand(cmds...)

	return cmd
}

func keysCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "List",
			Short:   "List all keys in the keybase",
			Long:    "List all of the hex addresses of the keys stored in the keybase",
			Aliases: []string{"list"},
			Args:    cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Open the debug keybase at the specified path
				pocketDir := strings.TrimSuffix(dataDir, "/")
				keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
				if err != nil {
					return err
				}
				kb, err := keybase.NewKeybase(keybasePath)
				if err != nil {
					return err
				}

				addresses, _, err := kb.GetAll()
				if err != nil {
					return err
				}

				for _, addr := range addresses {
					fmt.Println(addr)
				}

				return nil
			},
		},
	}
	return cmds
}
