package cli

import (
	"fmt"

	"github.com/pokt-network/pocket/rpc"
	"github.com/spf13/cobra"
)

func init() {
	consensusCmd := NewConsensusCommand()
	rootCmd.AddCommand(consensusCmd)
}

func NewConsensusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Consensus",
		Short:   "Consensus specific commands",
		Aliases: []string{"consensus"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(consensusCommands()...)

	return cmd
}

func consensusCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "State",
			Short:   "Returns \"Height/Round/Step\"",
			Long:    "State returns the height, round and step in \"Height/Round/Step\" format",
			Aliases: []string{"state"},
			RunE: func(cmd *cobra.Command, args []string) error {
				response, err := getConsensusState(cmd)
				if err != nil {
					return err
				}

				fmt.Printf("%d/%d/%d\n", response.JSONDefault.Height, response.JSONDefault.Round, response.JSONDefault.Step)

				return nil
			},
		},
		{
			Use:     "Height",
			Short:   "Returns the Height",
			Long:    "Height returns the height in the node's current consensus state",
			Aliases: []string{"height"},
			RunE: func(cmd *cobra.Command, args []string) error {
				response, err := getConsensusState(cmd)
				if err != nil {
					return err
				}

				fmt.Printf("%d\n", response.JSONDefault.Height)

				return nil
			},
		},
		{
			Use:     "Round",
			Short:   "Returns the Round",
			Long:    "Round returns the round in the node's current consensus state",
			Aliases: []string{"round"},
			RunE: func(cmd *cobra.Command, args []string) error {
				response, err := getConsensusState(cmd)
				if err != nil {
					return err
				}

				fmt.Printf("%d\n", response.JSONDefault.Round)

				return nil
			},
		},
		{
			Use:     "Step",
			Short:   "Returns the Step",
			Long:    "Step returns the step in the node's current consensus state",
			Aliases: []string{"step"},
			RunE: func(cmd *cobra.Command, args []string) error {
				response, err := getConsensusState(cmd)
				if err != nil {
					return err
				}

				fmt.Printf("%d\n", response.JSONDefault.Step)

				return nil
			},
		},
	}
	return cmds
}

func getConsensusState(cmd *cobra.Command) (*rpc.GetV1ConsensusStateResponse, error) {
	client, err := rpc.NewClientWithResponses(remoteCLIURL)
	if err != nil {
		return nil, nil
	}
	response, err := client.GetV1ConsensusStateWithResponse(cmd.Context())
	if err != nil {
		return nil, unableToConnectToRpc(err)
	}
	return response, nil
}
