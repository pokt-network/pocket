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

	cmd.AddCommand(govCommands()...)

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
				// TODO(deblasis): implement RPC client, route and handler
				fmt.Printf("changing parameter %s owned by %s to %s\n", args[1], args[0], args[2])

				// TODO(pocket/issues/150): update when we have keybase
				pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
				if err != nil {
					return err
				}

				key := args[1]
				value := args[2]

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
