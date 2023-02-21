package cli

import (
	"fmt"
	"math/big"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pokt-network/pocket/app/client/keybase"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
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
		ActorType coreTypes.ActorType
		Options   []cmdOption
	}
)

func NewActorCommands(cmdOptions []cmdOption) []*cobra.Command {
	actorCmdDefs := []actorCmdDef{
		{"Application", coreTypes.ActorType_ACTOR_TYPE_APP, cmdOptions},
		{"Servicer", coreTypes.ActorType_ACTOR_TYPE_SERVICER, cmdOptions},
		{"Fisherman", coreTypes.ActorType_ACTOR_TYPE_FISH, cmdOptions},
		{"Validator", coreTypes.ActorType_ACTOR_TYPE_VAL, cmdOptions},
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
	applySubcommandOptions(cmds, cmdDef.Options)
	return cmds
}

func newStakeCmd(cmdDef actorCmdDef) *cobra.Command {
	short := fmt.Sprintf("Stake a %s in the network. Custodial stake uses the same address as operator/output for rewards/return of staked funds.", cmdDef.Name)
	long := fmt.Sprintf(`Stake the %s into the network, making it available for service.

Will prompt the user for the *fromAddr* account passphrase. If the %s is already staked, this transaction acts as an *update* transaction.

A %s can update relayChainIDs, serviceURI, and raise the stake amount with this transaction.

If the %s is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account.

If no changes are desired for the parameter, just enter the current param value just as before.`, cmdDef.Name, cmdDef.Name, cmdDef.Name, cmdDef.Name)
	stakeCmd := &cobra.Command{
		Use:   "Stake <fromAddr> <amount> <relayChainIDs> <serviceURI>",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Unpack CLI arguments
			fromAddrHex := args[0]
			amount := args[1]

			// Open the keybase at the specified directory
			pocketDir := strings.TrimSuffix(dataDir, "/")
			keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
			if err != nil {
				return err
			}
			kb, err := keybase.NewKeybase(keybasePath)
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

			err = validateStakeAmount(amount)
			if err != nil {
				return err
			}

			// removing all invalid characters from rawChains argument
			rawChains := rawChainCleanupRegex.ReplaceAllString(args[2], "")
			chains := strings.Split(rawChains, ",")
			serviceURI := args[3]

			msg := &typesUtil.MessageStake{
				PublicKey:     pk.PublicKey().Bytes(),
				Chains:        chains,
				Amount:        amount,
				ServiceUrl:    serviceURI,
				OutputAddress: pk.Address(),
				Signer:        pk.Address(),
				ActorType:     cmdDef.ActorType,
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
	}

	return stakeCmd
}

func newEditStakeCmd(cmdDef actorCmdDef) *cobra.Command {
	editStakeCmd := &cobra.Command{
		Use:   "EditStake <fromAddr> <amount> <relayChainIDs> <serviceURI>",
		Short: "EditStake <fromAddr> <amount> <relayChainIDs> <serviceURI>",
		Long:  fmt.Sprintf(`Stakes a new <amount> for the %s actor with address <fromAddr> for the specified <relayChainIDs> and <serviceURI>.`, cmdDef.Name),
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Unpack CLI arguments
			fromAddrHex := args[0]
			fromAddr := crypto.AddressFromString(args[0])
			amount := args[1]

			// Open the keybase at the specified directory
			pocketDir := strings.TrimSuffix(dataDir, "/")
			keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
			if err != nil {
				return err
			}
			kb, err := keybase.NewKeybase(keybasePath)
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

			err = validateStakeAmount(amount)
			if err != nil {
				return err
			}
			// removing all invalid characters from rawChains argument
			rawChains := rawChainCleanupRegex.ReplaceAllString(args[2], "")
			chains := strings.Split(rawChains, ",")
			serviceURI := args[3]

			msg := &typesUtil.MessageEditStake{
				Address:    fromAddr,
				Chains:     chains,
				Amount:     amount,
				ServiceUrl: serviceURI,
				Signer:     pk.Address(),
				ActorType:  cmdDef.ActorType,
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
	}
	return editStakeCmd
}

func newUnstakeCmd(cmdDef actorCmdDef) *cobra.Command {
	unstakeCmd := &cobra.Command{
		Use:   "Unstake <fromAddr>",
		Short: "Unstake <fromAddr>",
		Long:  fmt.Sprintf(`Unstakes the previously staked tokens for the %s actor with address <fromAddr>`, cmdDef.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Unpack CLI arguments
			fromAddrHex := args[0]

			// Open the keybase at the specified directory
			pocketDir := strings.TrimSuffix(dataDir, "/")
			keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
			if err != nil {
				return err
			}
			kb, err := keybase.NewKeybase(keybasePath)
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

			msg := &typesUtil.MessageUnstake{
				Address:   pk.Address(),
				Signer:    pk.Address(),
				ActorType: cmdDef.ActorType,
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
			// Unpack CLI arguments
			fromAddrHex := args[0]

			// Open the keybase at the specified directory
			pocketDir := strings.TrimSuffix(dataDir, "/")
			keybasePath, err := filepath.Abs(pocketDir + keybaseSuffix)
			if err != nil {
				return err
			}
			kb, err := keybase.NewKeybase(keybasePath)
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

			msg := &typesUtil.MessageUnpause{
				Address:   pk.Address(),
				Signer:    pk.Address(),
				ActorType: cmdDef.ActorType,
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
	}
	return unpauseCmd
}
