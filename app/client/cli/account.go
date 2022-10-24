package cli

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
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
				// TODO(#150): update when we have keybase
				pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
				if err != nil {
					return err
				}
				// NOTE: since we don't have a keybase yet (tracked in #150), we are currently inferring the `fromAddr` from the PrivateKey supplied via the flag `--path_to_private_key_file`
				// the following line is commented out to show that once we have a keybase, `fromAddr` should come from the command arguments and not the PrivateKey (pk) anymore.
				//
				// fromAddr := crypto.AddressFromString(args[0])
				toAddr := crypto.AddressFromString(args[1])
				amount := args[2]

				msg := &types.MessageSend{
					FromAddress: pk.Address(),
					ToAddress:   toAddr,
					Amount:      amount,
				}

				tx, err := prepareTxJson(msg, pk)
				if err != nil {
					return err
				}

				resp, err := postRawTx(cmd.Context(), pk, tx)
				if err != nil {
					return err
				}
				// DISCUSS(#310): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
				fmt.Println(resp)

				return nil
			},
		},
	}
	return cmds
}
