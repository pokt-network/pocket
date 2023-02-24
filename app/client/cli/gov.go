package cli

import (
	"fmt"

	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func init() {
	rootCmd.AddCommand(NewGovernanceCommand())
}

func NewGovernanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Governance",
		Short:   "Governance specific commands",
		Aliases: []string{"gov"},
		Args:    cobra.ExactArgs(0),
	}

	cmds := govCommands()
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	cmd.AddCommand(cmds...)

	return cmd
}

func govCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "ChangeParameter <owner> <key> <value>",
			Short:   "ChangeParameter <owner> <key> <value>",
			Long:    "Changes the Governance parameter with <key> owned by <owner> to <value>",
			Aliases: []string{},
			Args:    cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI arguments
				fromAddrHex := args[0]
				key := args[1]
				value := args[2]

				// TODO(deblasis): implement RPC client, route and handler
				fmt.Printf("changing parameter %s owned by %s to %s\n", args[1], args[0], args[2])

				kb, err := keybaseForCli()
				if err != nil {
					return err
				}

				pwd = readPassphrase(pwd)

				pk, err := kb.GetPrivKey(fromAddrHex, pwd)
				if err != nil {
					return err
				}
				if err := kb.Stop(); err != nil {
					return err
				}

				pbValue, err := anypb.New(wrapperspb.String(value))
				if err != nil {
					return err
				}

				msg := &types.MessageChangeParameter{
					Signer:         pk.Address(),
					Owner:          pk.Address(),
					ParameterKey:   key,
					ParameterValue: pbValue,
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
