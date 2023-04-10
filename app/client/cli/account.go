package cli

import (
	"fmt"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewAccountCommand())
}

func NewAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Account",
		Short:   "Account specific commands",
		Aliases: []string{"account"},
		Args:    cobra.ExactArgs(0),
	}

	cmds := accountCommands()
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	cmd.AddCommand(cmds...)

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
				// Unpack CLI arguments
				fromAddrHex := args[0]
				fromAddr := crypto.AddressFromString(args[0])
				toAddr := crypto.AddressFromString(args[1])
				amount := args[2]

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !nonInteractive {
					pwd = readPassphrase(pwd)
				}

				pk, err := kb.GetPrivKey(fromAddrHex, pwd)
				if err != nil {
					return err
				}
				if err := kb.Stop(); err != nil {
					return err
				}

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
