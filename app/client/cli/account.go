package cli

import (
	"fmt"
	"github.com/pokt-network/pocket/app/client/keybase/debug"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	accountCmd := NewAccountCommand()
	rootCmd.AddCommand(accountCmd)
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
				// Open the debug keybase at $HOME/.pocket/keys
				kb, err := debug.NewDebugKeybase()
				if err != nil {
					return err
				}

				pwd = readPassphrase(pwd)

				fromAddrHex := args[0]
				pk, err := kb.GetPrivKey(fromAddrHex, pwd)
				if err != nil {
					return err
				}
				if err := kb.Stop(); err != nil {
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
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	return cmds
}
