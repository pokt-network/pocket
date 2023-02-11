package cli

import (
	"fmt"
	"path/filepath"

	"github.com/pokt-network/pocket/app/client/keybase"
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

				keybaseDir, err := filepath.Abs(dataDir + "/keys")
				if err != nil {
					return err
				}

				kb, err := keybase.InitialiseKeybase(keybaseDir)
				if err != nil {
					return err
				}

				pwd = readPassphrase(pwd)

				pk, err := kb.GetPrivKey(args[0], pwd)
				if err != nil {
					return err
				}
				if err := kb.Stop(); err != nil {
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
	for _, cmd := range cmds {
		cmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	}
	return cmds
}
