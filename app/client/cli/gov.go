package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/utility/types"
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
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
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

				// TODO(0xbigboss): implement RPC client, route and handler
				fmt.Printf("changing parameter %s owned by %s to %s\n", args[1], args[0], args[2])

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !flags.NonInteractive {
					pwd = readPassphrase(pwd)
				}

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

				if resp.StatusCode() != 200 {
					return fmt.Errorf("HTTP status code: %d\n", resp.StatusCode())
				}

				fmt.Printf("Successfully sent change parameter %s owned by %s to %s\n", args[1], args[0], args[2])
				fmt.Printf("HTTP status code: %d\n", resp.StatusCode())
				fmt.Println(string(resp.Body))

				return nil
			},
		},
		{
			Use:     "Upgrade <owner> <version> <height>",
			Short:   "Upgrade ",
			Long:    "Schedules an upgrade to the specified version at the specified height",
			Aliases: []string{"upgrade"},
			Args:    cobra.ExactArgs(3),
			RunE: func(cmd *cobra.Command, args []string) error {
				// Unpack CLI arguments
				fromAddrHex := args[0]
				version := args[1]
				heightArg := args[2]

				fmt.Printf("submitting upgrade for version %s at height %s.\n", args[0], args[1])

				kb, err := keybaseForCLI()
				if err != nil {
					return err
				}

				if !flags.NonInteractive {
					pwd = readPassphrase(pwd)
				}

				pk, err := kb.GetPrivKey(fromAddrHex, pwd)
				if err != nil {
					return err
				}
				if err := kb.Stop(); err != nil {
					return err
				}

				height, err := strconv.ParseInt(heightArg, 10, 64)
				if err != nil {
					return err
				}

				// TODO(0xbigboss): be nice and validate the inputs before submitting, instead of reverting tx.

				msg := &types.MessageUpgrade{
					Signer:  pk.Address(),
					Version: version,
					Height:  height,
				}

				err = msg.ValidateBasic()
				if err != nil {
					cmd.PrintErrf("invalid message: %s\n", err)
					return err
				}

				tx, err := prepareTxBytes(msg, pk)
				if err != nil {
					return err
				}

				resp, err := postRawTx(cmd.Context(), pk, tx)
				if err != nil {
					return err
				}

				if resp.StatusCode() != 200 {
					return fmt.Errorf("HTTP status code: %d\n", resp.StatusCode())
				}

				fmt.Printf("Successfully submitted upgrade for version %s at height %s.\n", args[0], args[1])
				fmt.Printf("HTTP status code: %d\n", resp.StatusCode())
				fmt.Println(string(resp.Body))

				return nil
			},
		},
		// TODO: 0xbigboss MessageCancelUpgrade
	}
	return cmds
}
