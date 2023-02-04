package cli

import (
	"fmt"
	"os"

	"github.com/pokt-network/pocket/app/client/keybase"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

// Hardcoded keybase related constants
const (
	KEYBASE_PATH_SUFFIX    = "/.pocket/keys"      // TODO: Find a good place for this
	PRIVATEKEY_YAML_SUFFIX = "/private-keys.yaml" // Remove when PR#354 is merged then use `build/localnet/manifests/private-keys.yaml`
)

func init() {
	accounCmd := NewAccountCommand()
	accounCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	rootCmd.AddCommand(accounCmd)
}

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

func accountCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Send <fromAddr> <to> <amount>",
			Short:   "Send <fromAddr> <to> <amount>",
			Long:    "Sends <amount> to address <to> from address <fromAddr>",
			Aliases: []string{"send"},
			Args:    cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				keybase, err := keybase.InitialiseKeybase(homeDir+KEYBASE_PATH_SUFFIX, homeDir+PRIVATEKEY_YAML_SUFFIX) // Change when PR#354 is merged
				if err != nil {
					return err
				}

				// TODO (team): passphrase is currently not used since there's no keybase yet, the prompt is here to mimick the real world UX
				pwd = readPassphrase(pwd)

				pk, err := keybase.GetPrivKey(args[0], pwd)
				if err != nil {
					return err
				}

				fromAddr := crypto.AddressFromString(args[0])
				toAddr := crypto.AddressFromString(args[1])
				amount := args[2]

				msg := &types.MessageSend{
					FromAddress: fromAddr,
					ToAddress:   toAddr,
					Amount:      amount,
				}

				tx, err := prepareTxBytes(msg, pk)
				if err != nil {
					return err
				}

				resp, err := postRawTx(cmd.Context(), pk, tx)
				if err != nil {
					return err
				}

				// DISCUSS(#310): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
				fmt.Printf("HTTP status code: %d\n", resp.StatusCode())
				fmt.Println(string(resp.Body))

				return nil
			},
		},
	}
	return cmds
}
