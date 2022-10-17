package cli

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	actorCmd := NewAccountCommand()
	actorCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	rootCmd.AddCommand(actorCmd)
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
				// TODO(pocket/issues/150): update when we have keybase
				pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
				if err != nil {
					return err
				}
				// currently ignored since we are using the address from the PrivateKey
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
				// DISCUSS(team): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
				fmt.Println(resp)

				return nil
			},
		},
	}
	return cmds
}
