package cli

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	typesPersistence "github.com/pokt-network/pocket/persistence/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewActorCommands(attachPwdFlagToSubcommands())...)

	rawChainCleanupRegex = regexp.MustCompile(rawChainCleanupExpr)
}

const (
	// DISCUSS(team): this should probably come from somewhere else
	stakingRecommendationAmount = 15100000000
	rawChainCleanupExpr         = "[^,a-fA-F0-9]+"
)

var (
	pwd                  string
	rawChainCleanupRegex *regexp.Regexp
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

	cmds := make([]*cobra.Command, 0)
	for _, cmdDef := range actorCmdDefs {
		cmd := &cobra.Command{
			Use:     cmdDef.Name,
			Short:   fmt.Sprintf("%s actor specific commands", cmdDef.Name),
			Aliases: []string{strings.ToLower(cmdDef.Name), typesUtil.ActorType_name[int32(cmdDef.ActorType)]}, // TODO: create some helper function to convert enum to names and viceversa
			Args:    cobra.ExactArgs(0),
		}
		cmd.AddCommand(newActorCommands(cmdDef)...)
		cmds = append(cmds, cmd)
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
		Use:   "Custodial <fromAddr> <amount> <RelayChainIDs> <serviceURI>",
		Short: "Stake a node in the network. Custodial stake uses the same address as operator/output for rewards/return of staked funds.",
		Long: `Stake the node into the network, making it available for service.
Will prompt the user for the <fromAddr> account passphrase. If the node is already staked, this transaction acts as an *update* transaction.
A node can update relayChainIDs, serviceURI, and raise the stake amount with this transaction.
If the node is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account
If no changes are desired for the parameter, just enter the current param value just as before.`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}
			// currently ignored since we are using the address from the PrivateKey
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

			j, err := prepareTx(msg, pk)
			if err != nil {
				return err
			}

			resp, err := postRawTx(cmd.Context(), pk, j)
			if err != nil {
				return err
			}
			// DISCUSS(team): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
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
		Use:   "EditStake <fromAddr> <amount> <RelayChainIDs> <serviceURI>",
		Short: "EditStake <fromAddr> <amount> <RelayChainIDs> <serviceURI>",
		Long:  fmt.Sprintf(`Stakes a new <amount> for the %s actor with address <fromAddr> for the specified <RelayChainIDs> and <serviceURI>.`, cmdDef.Name),
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
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

			j, err := prepareTx(msg, pk)
			if err != nil {
				return err
			}

			resp, err := postRawTx(cmd.Context(), pk, j)
			if err != nil {
				return err
			}
			// DISCUSS(team): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
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

			j, err := prepareTx(msg, pk)
			if err != nil {
				return err
			}

			resp, err := postRawTx(cmd.Context(), pk, j)
			if err != nil {
				return err
			}
			// DISCUSS(team): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
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
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
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

			j, err := prepareTx(msg, pk)
			if err != nil {
				return err
			}

			resp, err := postRawTx(cmd.Context(), pk, j)
			if err != nil {
				return err
			}
			// DISCUSS(team): define UX for return values - should we return the raw response or a parsed/human readable response? For now, I am simply printing to stdout
			fmt.Println(resp)

			return nil
		},
	}
	return unpauseCmd
}

func readPassphrase(currPwd string) string {
	if strings.TrimSpace(currPwd) == "" {
		fmt.Println("Enter Passphrase: ")
	} else {
		fmt.Println("Using Passphrase provided via flag")
	}

	return Credentials(currPwd)
}

func validateStakeAmount(amount string) error {
	am, err := typesPersistence.StringToBigInt(amount)
	if err != nil {
		return err
	}

	sr := big.NewInt(stakingRecommendationAmount)
	if typesUtil.BigIntLessThan(am, sr) {
		fmt.Printf("The amount you are staking for is below the recommendation of %d POKT, would you still like to continue? y|n\n", sr.Div(sr, big.NewInt(1000000)).Int64())
		if !Confirmation(pwd) {
			return fmt.Errorf("aborted")
		}
	}
	return nil
}

func applySubcommandOptions(cmds []*cobra.Command, cmdDef actorCmdDef) {
	for _, cmd := range cmds {
		for _, opt := range cmdDef.Options {
			opt(cmd)
		}
	}
}

func attachPwdFlagToSubcommands() []cmdOption {
	return []cmdOption{func(c *cobra.Command) {
		c.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	}}
}
