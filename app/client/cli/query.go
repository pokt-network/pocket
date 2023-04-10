package cli

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/rpc"
	"github.com/spf13/cobra"
)

func init() {
	queryCmd := NewQueryCommand()
	rootCmd.AddCommand(queryCmd)
}

func NewQueryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Query",
		Short:   "Commands related to querying on-chain data via the node's RPC server",
		Aliases: []string{"query"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(queryCommands()...)

	return cmd
}

// TODO: (h5law) Split this into multiple functions to assign the relevent flags in batches
func queryCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Account <address> [--height]",
			Short:   "Get the accound data of an address at a specified height",
			Long:    "Queries the node RPC to obtain the account data of the speicifed account at the given height",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"account"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryAddressHeight{
					Address: args[1],
					Height:  0, // TODO: Change when add flag
				}

				response, err := client.PostV1QueryAccount(cmd.Context(), body)
				if err != nil {
					return unableToConnectToRpc(err)
				}
				statusCode := response.StatusCode
				resp, err := io.ReadAll(response.Body)
				if err != nil {
					logger.Global.Error().Err(err).Msg("Error reading response body")
					return err
				}
				if statusCode == http.StatusOK {
					fmt.Println(string(resp))
					return nil
				}
				return rpcResponseCodeUnhealthy(statusCode, resp)
			},
		},
		{
			Use:     "AllChainParams",
			Short:   "Get current values of all node parameters",
			Long:    "Queries the node RPC to obtain the current values of all the governance parameters",
			Aliases: []string{"allparams"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}
				response, err := client.GetV1QueryAllChainParams(cmd.Context())
				if err != nil {
					return unableToConnectToRpc(err)
				}
				statusCode := response.StatusCode
				body, err := io.ReadAll(response.Body)
				if err != nil {
					logger.Global.Error().Err(err).Msg("Error reading response body")
					return err
				}
				if statusCode == http.StatusOK {
					fmt.Println(string(body))
					return nil
				}
				return rpcResponseCodeUnhealthy(statusCode, body)
			},
		},
	}
	return cmds
}
