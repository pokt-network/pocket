package cli

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/rpc"
	"github.com/spf13/cobra"
)

var (
	height   int64
	page     int64
	per_page int64
	sort     string
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

	heightCmds := queryHeightCommands()
	heightPaginatedCmds := queryHeightPaginatedCommands()
	addressPaginatedCmds := queryAddressPaginatedCommands()
	getCmds := queryCommands()

	// attach --height flag
	applySubcommandOptions(heightCmds, attachHeightFlagToSubcommands())
	applySubcommandOptions(heightPaginatedCmds, attachHeightFlagToSubcommands())

	// attach --page, --per_page flags
	applySubcommandOptions(heightPaginatedCmds, attachPaginationFlagsToSubcommands())
	applySubcommandOptions(addressPaginatedCmds, attachPaginationFlagsToSubcommands())

	// attach --sort flag
	applySubcommandOptions(addressPaginatedCmds, attachSortFlagToSubcommands())

	cmd.AddCommand(heightCmds...)
	cmd.AddCommand(heightPaginatedCmds...)
	cmd.AddCommand(addressPaginatedCmds...)
	cmd.AddCommand(getCmds...)

	return cmd
}

func queryHeightCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Account <address> [--height]",
			Short:   "Get the account data of an address at a specified height",
			Long:    "Queries the node RPC to obtain the account data of the speicifed account at the given height",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"account"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryAddressHeight{
					Address: args[0],
					Height:  height,
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
			Use:     "Balance <address> [--height]",
			Short:   "Get the balance of an address at a specified height",
			Long:    "Queries the node RPC to obtain the balance of the account at the given height",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"balance"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryAddressHeight{
					Address: args[0],
					Height:  height,
				}

				response, err := client.PostV1QueryBalance(cmd.Context(), body)
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
			Use:     "Block [--height]",
			Short:   "Get the block data of the specified height",
			Long:    "Queries the node RPC to obtain the block data at the given height",
			Args:    cobra.ExactArgs(0),
			Aliases: []string{"block"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryHeight{
					Height: height,
				}

				response, err := client.PostV1QueryBlock(cmd.Context(), body)
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
	}

	return cmds
}

func queryHeightPaginatedCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Accounts [--height] [--page] [--per_page]",
			Short:   "Get the account data of all accounts the specified height",
			Long:    "Queries the node RPC to obtain the paginated data for all accounts at the given height",
			Args:    cobra.ExactArgs(0),
			Aliases: []string{"accounts"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryHeightPaginated{
					Height:  height,
					Page:    page,
					PerPage: per_page,
				}

				response, err := client.PostV1QueryAccounts(cmd.Context(), body)
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
	}

	return cmds
}

func queryAddressPaginatedCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "AccountTxs <address> [--page] [--per_page] [--sort]",
			Short:   "Get all the transaction data of the given address",
			Long:    "Queries the node RPC to obtain the paginated data for all transactions from the given address",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"accounttxs"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryAddressPaginated{
					Address: args[0],
					Page:    page,
					PerPage: per_page,
					Sort:    &sort,
				}

				response, err := client.PostV1QueryAccounttxs(cmd.Context(), body)
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
	}

	return cmds
}

func queryCommands() []*cobra.Command {
	cmds := []*cobra.Command{
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
