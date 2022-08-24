package cli

import (
	"fmt"
	"strings"

	"github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewActorCommands()...)
}

type actorCmdDef struct {
	Name      string
	ActorType types.ActorType
}

func NewActorCommands() (cmds []*cobra.Command) {
	actorCmdDefs := []actorCmdDef{
		{"Application", types.ActorType_App},
		{"Node", types.ActorType_Node},
		{"Fisherman", types.ActorType_Fish},
		{"Validator", types.ActorType_Val},
	}

	for _, cmdDef := range actorCmdDefs {

		cmd := &cobra.Command{
			Use:     cmdDef.Name,
			Short:   fmt.Sprintf("%s actor specific commands", cmdDef.Name),
			Aliases: []string{strings.ToLower(cmdDef.Name), cmdDef.ActorType.GetActorName()},
			Args:    cobra.ExactArgs(0),
		}
		cmd.AddCommand(newActorCommands(cmdDef)...)
		cmds = append(cmds, cmd)
	}

	return cmds
}

func newActorCommands(cmdDef actorCmdDef) (cmds []*cobra.Command) {
	stakeCmd := &cobra.Command{
		Use:   "Stake <from> <amount>",
		Short: "Stake <from> <amount>",
		Long:  fmt.Sprintf(`Stakes <amount> for the %s actot with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}
			// currently ignored since we are using the address from the PrivateKey
			// fromAddr := crypto.AddressFromString(args[0])
			amount := args[1]

			_ = &types.MessageStake{
				PublicKey:     pk.PublicKey().Bytes(),
				Chains:        []string{}, // TODO(deblasis): 👀
				Amount:        amount,
				ServiceUrl:    "",       // TODO(deblasis): 👀
				OutputAddress: []byte{}, // TODO(deblasis): 👀
				Signer:        []byte{}, // TODO(deblasis): 👀 pk.Address() ?
				ActorType:     cmdDef.ActorType,
			}

			return nil
		},
	}
	cmds = append(cmds, stakeCmd)

	editStakeCmd := &cobra.Command{
		Use:   "EditStake <from> <amount>",
		Short: "EditStake <from> <amount>",
		Long:  fmt.Sprintf(`Stakes a new <amount> for the %s actor with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}

			amount := args[1]

			_ = &types.MessageEditStake{
				Address:    pk.Address(),
				Chains:     []string{}, // TODO(deblasis): 👀
				Amount:     amount,
				ServiceUrl: "",       // TODO(deblasis): 👀
				Signer:     []byte{}, // TODO(deblasis): 👀
				ActorType:  cmdDef.ActorType,
			}

			return nil
		},
	}
	cmds = append(cmds, editStakeCmd)

	unstakeCmd := &cobra.Command{
		Use:   "Unstake <from>",
		Short: "Unstake <from>",
		Long:  fmt.Sprintf(`Unstakes the prevously staked tokens for the %s actor with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}

			_ = &types.MessageUnstake{
				Address:   pk.Address(),
				Signer:    []byte{}, // TODO(deblasis): 👀
				ActorType: cmdDef.ActorType,
			}

			return nil
		},
	}
	cmds = append(cmds, unstakeCmd)

	unpauseCmd := &cobra.Command{
		Use:   "Unpause <from>",
		Short: "Unpause <from>",
		Long:  fmt.Sprintf(`Unpauses the %s actor with address <from>`, cmdDef.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(pocket/issues/150): update when we have keybase
			pk, err := readEd25519PrivateKeyFromFile(privateKeyFilePath)
			if err != nil {
				return err
			}

			_ = &types.MessageUnpause{
				Address:   pk.Address(),
				Signer:    []byte{}, // TODO(deblasis): 👀
				ActorType: cmdDef.ActorType,
			}

			return nil
		},
	}
	cmds = append(cmds, unpauseCmd)

	return cmds
}
