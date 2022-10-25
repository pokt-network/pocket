package cli

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewActorCommands(attachPwdFlagToSubcommands())...)

	rawChainCleanupRegex = regexp.MustCompile(rawChainCleanupExpr)

	oneMillion = big.NewInt(1e6)
}

const (
	// DISCUSS(team): this should probably come from somewhere else
	stakingRecommendationAmount = 15100000000
	rawChainCleanupExpr         = "[^,a-fA-F0-9]+"
)

var (
	pwd                  string
	rawChainCleanupRegex *regexp.Regexp
	oneMillion           *big.Int
)

type (
	cmdOption   func(*cobra.Command)
	actorCmdDef struct {
		Name      string
		ActorType typesUtil.ActorType
		Options   []cmdOption
	}
)

func NewActorCommands(cmdOptions []cmdOption) []*cobra.Command {
	actorCmdDefs := []actorCmdDef{
		{"Application", typesUtil.ActorType_App, cmdOptions},
		{"Node", typesUtil.ActorType_ServiceNode, cmdOptions},
		{"Fisherman", typesUtil.ActorType_Fisherman, cmdOptions},
		{"Validator", typesUtil.ActorType_Validator, cmdOptions},
	}

	cmds := make([]*cobra.Command, len(actorCmdDefs))
	for i, cmdDef := range actorCmdDefs {
		cmd := &cobra.Command{
			Use:     cmdDef.Name,
			Short:   fmt.Sprintf("%s actor specific commands", cmdDef.Name),
			Aliases: []string{strings.ToLower(cmdDef.Name), cmdDef.ActorType.GetName()},
			Args:    cobra.ExactArgs(0),
		}
		cmd.AddCommand(newActorCommands(cmdDef)...)
		cmds[i] = cmd
	}
	return cmds
}

func newActorCommands(cmdDef actorCmdDef) []*cobra.Command {
	cmds := []*cobra.Command{
		newStakeCmd(cmdDef),
		newEditStakeCmd(cmdDef),
		newUnstakeCmd(cmdDef),
		newUnpauseCmd(cmdDef),
	}
	applySubcommandOptions(cmds, cmdDef)
	return cmds
}

func newStakeCmd(cmdDef actorCmdDef) *cobra.Command {
	stakeCmd := &cobra.Command{
		Use:   "Stake",
		Short: fmt.Sprintf("Stake an actor (%s) in the network.", cmdDef.Name),
		Long:  fmt.Sprintf("Stake the %s actor into the network, making it available for service.", cmdDef.Name),
	}

	custodialStakeCmd := &cobra.Command{
		Use:   "Custodial <fromAddr> <amount> <relayChainIDs> <serviceURI>",
		Short: "Stake a node in the network. Custodial stake uses the same address as operator/output for rewards/return of staked funds.",
		Long: `Stake the node into the network, making it available for service.
Will prompt the user for the <fromAddr> account passphrase. If the node is already staked, this transaction acts as an *update* transaction.
A node can update relayChainIDs, serviceURI, and raise the stake amount with this transaction.
If the node is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account
If no changes are desired for the parameter, just enter the current param value just as before.`,
		Args: cobra.ExactArgs(4), // REFACTOR(#150): <fromAddr> not being used at the moment. Update once a keybase is implemented.
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
			amount := args[1]
			err = validateStakeAmount(amount)
			if err != nil {
				return err
			}

			// removing all invalid characters from rawChains argument
			rawChains := rawChainCleanupRegex.ReplaceAllString(args[2], "")
			chains := strings.Split(rawChains, ",")
			serviceURI := args[3]

			// TODO (team): passphrase is currently not used since there's no keybase yet, the prompt is here to mimick the real world UX
			pwd = readPassphrase(pwd)

			msg := &typesUtil.MessageStake{
				PublicKey:     pk.PublicKey().Bytes(),
				Chains:        chains,
				Amount:        amount,
				ServiceUrl:    serviceURI,
				OutputAddress: pk.Address(),
				Signer:        pk.Address(),
				ActorType:     cmdDef.ActorType,
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
	}

	stakeCmd.AddCommand(custodialStakeCmd)

	applySubcommandOptions(stakeCmd.Commands(), cmdDef)

	return stakeCmd
}

func newEditStakeCmd(cmdDef actorCmdDef) *cobra.Command {
	editStakeCmd := &cobra.Command{
		Use:   "EditStake <fromAddr> <amount> <relayChainIDs> <serviceURI>",
		Short: "EditStake <fromAddr> <amount> <relayChainIDs> <serviceURI>",
		Long:  fmt.Sprintf(`Stakes a new <amount> for the %s actor with address <fromAddr> for the specified <relayChainIDs> and <serviceURI>.`, cmdDef.Name),
		Args:  cobra.ExactArgs(4), // REFACTOR(#150): <fromAddr> not being used at the moment. Update once a keybase is implemented.
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}

			amount := args[1]
			err = validateStakeAmount(amount)
			if err != nil {
				return err
			}
			// removing all invalid characters from rawChains argument
			rawChains := rawChainCleanupRegex.ReplaceAllString(args[2], "")
			chains := strings.Split(rawChains, ",")
			serviceURI := args[3]

			// TODO (team): passphrase is currently not used since there's no keybase yet, the prompt is here to mimick the real world UX
			pwd = readPassphrase(pwd)

			msg := &typesUtil.MessageEditStake{
				Address:    pk.Address(),
				Chains:     chains,
				Amount:     amount,
				ServiceUrl: serviceURI,
				Signer:     pk.Address(),
				ActorType:  cmdDef.ActorType,
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
	}
	return editStakeCmd
}

func newUnstakeCmd(cmdDef actorCmdDef) *cobra.Command {
	unstakeCmd := &cobra.Command{
		Use:   "Unstake <fromAddr>",
		Short: "Unstake <fromAddr>",
		Long:  fmt.Sprintf(`Unstakes the prevously staked tokens for the %s actor with address <fromAddr>`, cmdDef.Name),
		Args:  cobra.ExactArgs(1), // REFACTOR(#150): <fromAddr> not being used at the moment. Update once a keybase is implemented.
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}

			// TODO (team): passphrase is currently not used since there's no keybase yet, the prompt is here to mimick the real world UX
			pwd = readPassphrase(pwd)

			msg := &typesUtil.MessageUnstake{
				Address:   pk.Address(),
				Signer:    pk.Address(),
				ActorType: cmdDef.ActorType,
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
	}
	return unstakeCmd
}

func newUnpauseCmd(cmdDef actorCmdDef) *cobra.Command {
	unpauseCmd := &cobra.Command{
		Use:   "Unpause <fromAddr>",
		Short: "Unpause <fromAddr>",
		Long:  fmt.Sprintf(`Unpauses the %s actor with address <fromAddr>`, cmdDef.Name),
		Args:  cobra.ExactArgs(1), // REFACTOR(#150): Not being used at the moment. Update once a keybase is implemented.
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(#150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}

			// TODO (team): passphrase is currently not used since there's no keybase yet, the prompt is here to mimick the real world UX
			pwd = readPassphrase(pwd)

			msg := &typesUtil.MessageUnpause{
				Address:   pk.Address(),
				Signer:    pk.Address(),
				ActorType: cmdDef.ActorType,
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
	}
	return unpauseCmd
}
