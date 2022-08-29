package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strings"

	"github.com/pokt-network/pocket/app/client/rpc"
	sharedTypes "github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility/types"
	utilityTypes "github.com/pokt-network/pocket/utility/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewActorCommands()...)
}

var (
	// DISCUSS(team): this should probably come from somewhere else
	stakingRecommendationAmount = big.NewInt(15100000000)
	pwd                         string
)

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
		Use:   "Stake <fromAddr> <amount> <RelayChainIDs> <serviceURI>",
		Short: "Stake a node in the network. Custodial stake uses the same address as operator/output for rewards/return of staked funds.",
		Long: `Stake the node into the network, making it available for service.
Will prompt the user for the <fromAddr> account passphrase. If the node is already staked, this transaction acts as an *update* transaction.
A node can updated relayChainIDs, serviceURI, and raise the stake amount with this transaction.
If the node is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account
If no changes are desired for the parameter, just enter the current param value just as before`,
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

			am, err := sharedTypes.StringToBigInt(args[1])
			if err != nil {
				log.Fatal(err)
			}

			if sharedTypes.BigIntLessThan(am, stakingRecommendationAmount) {
				fmt.Printf("The amount you are staking for is below the recommendation of %d POKT, would you still like to continue? y|n\n", stakingRecommendationAmount.Div(stakingRecommendationAmount, big.NewInt(1000000)).Int64())
				if !Confirmation(pwd) {
					return fmt.Errorf("aborted")
				}
			}

			reg, err := regexp.Compile("[^,a-fA-F0-9]+")
			if err != nil {
				log.Fatal(err)
			}
			rawChains := reg.ReplaceAllString(args[2], "")
			chains := strings.Split(rawChains, ",")
			serviceURI := args[3]

			fmt.Println("Enter Passphrase: ")
			// TODO (team): passphrase is currently not used since there's no keybase yet, the prompt is here to mimick the real world UX
			_ = Credentials(pwd)

			msg := &types.MessageStake{
				PublicKey:     pk.PublicKey().Bytes(),
				Chains:        chains,
				Amount:        amount,
				ServiceUrl:    serviceURI,
				OutputAddress: []byte{},     // TODO(deblasis): 👀
				Signer:        pk.Address(), // TODO(deblasis): figure out where I should get this from (https://github.com/pokt-network/pocket/pull/169#discussion_r953186602)
				ActorType:     cmdDef.ActorType,
			}

			codec := sharedTypes.GetCodec()
			anyMsg, err := codec.ToAny(msg)

			signature, _ := pk.Sign(msg.GetSignBytes())

			tx := &utilityTypes.Transaction{
				Msg: anyMsg,
				Signature: &utilityTypes.Signature{
					Signature: signature,
					PublicKey: pk.PublicKey().Bytes(),
				},
				Nonce: "", // TODO(deblasis): figure out where I should get this from
			}

			j, err := json.Marshal(tx)
			if err != nil {
				return err
			}

			// TODO(deblasis): we need a single source of truth for routes, the empty string should be replaced with something like a constant that can be used to point to a specific route
			// perhaps the routes could be centralized into a map[string]Route in #176 and accessed here
			// I will do this in #169 since it has commits from #176 and #177
			resp, err := QueryRPC(rpc.BroadcastTxSyncRoute, j)
			if err != nil {
				return err
			}
			fmt.Println(resp)

			return nil
		},
	}
	cmds = append(cmds, stakeCmd)

	editStakeCmd := &cobra.Command{
		Use:   "EditStake <fromAddr> <amount>",
		Short: "EditStake <fromAddr> <amount>",
		Long:  fmt.Sprintf(`Stakes a new <amount> for the %s actor with address <fromAddr>`, cmdDef.Name),
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
