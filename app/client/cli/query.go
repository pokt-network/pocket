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
	heightPaginatedSortedCmds := queryHeightPaginatedSortedCommands()
	paginatedSortedCmds := queryPaginatedSortedCommands()
	getCmds := queryCommands()

	// attach --height flag
	applySubcommandOptions(heightCmds, attachHeightFlagToSubcommands())
	applySubcommandOptions(heightPaginatedCmds, attachHeightFlagToSubcommands())
	applySubcommandOptions(heightPaginatedSortedCmds, attachHeightFlagToSubcommands())

	// attach --page, --per_page flags
	applySubcommandOptions(heightPaginatedCmds, attachPaginationFlagsToSubcommands())
	applySubcommandOptions(paginatedSortedCmds, attachPaginationFlagsToSubcommands())
	applySubcommandOptions(heightPaginatedSortedCmds, attachPaginationFlagsToSubcommands())

	// attach --sort flag
	applySubcommandOptions(paginatedSortedCmds, attachSortFlagToSubcommands())
	applySubcommandOptions(heightPaginatedSortedCmds, attachSortFlagToSubcommands())

	cmd.AddCommand(heightCmds...)
	cmd.AddCommand(heightPaginatedCmds...)
	cmd.AddCommand(heightPaginatedSortedCmds...)
	cmd.AddCommand(paginatedSortedCmds...)
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
			Use:     "App <address> [--height]",
			Short:   "Get the app data of an address at a specified height",
			Long:    "Queries the node RPC to obtain the app data of the speicifed address at the given height",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"app"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryAddressHeight{
					Address: args[0],
					Height:  height,
				}

				response, err := client.PostV1QueryApp(cmd.Context(), body)
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
		{
			Use:     "Param <parameter_name> [--height]",
			Short:   "Get the value of the parameter at the specified height",
			Long:    "Queries the node RPC to obtain the value of the specified parameter",
			Aliases: []string{"param"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryParameter{
					ParamName: args[0],
					Height:    height,
				}

				response, err := client.PostV1QueryParam(cmd.Context(), body)
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
			Use:     "Upgrade [--height]",
			Short:   "Get the upgrade version at the specified height",
			Long:    "Queries the node RPC to obtain the upgrade version for the specified height",
			Aliases: []string{"param"},
			Args:    cobra.ExactArgs(0),
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryHeight{
					Height: height,
				}

				response, err := client.PostV1QueryUpgrade(cmd.Context(), body)
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
		{
			Use:     "Apps [--height] [--page] [--per_page]",
			Short:   "Get all the data of all apps the specified height",
			Long:    "Queries the node RPC to obtain the paginated data for all apps at the given height",
			Args:    cobra.ExactArgs(0),
			Aliases: []string{"apps"},
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

				response, err := client.PostV1QueryApps(cmd.Context(), body)
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

func queryHeightPaginatedSortedCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "BlockTxs [--height] [--page] [--per_page] [--sort]",
			Short:   "Get all the transactions in the block with the specified height",
			Long:    "Queries the node RPC to obtain the paginated transactions in the block at the given height",
			Args:    cobra.ExactArgs(0),
			Aliases: []string{"blocktxs"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryHeightPaginated{
					Height:  height,
					Page:    page,
					PerPage: per_page,
					Sort:    &sort,
				}

				response, err := client.PostV1QueryBlocktxs(cmd.Context(), body)
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

func queryPaginatedSortedCommands() []*cobra.Command {
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
		{
			Use:     "UnconfirmedTxs [--page] [--per_page] [--sort]",
			Short:   "Get all the unconfirmed transaction data from the mempool",
			Long:    "Queries the node RPC to obtain the paginated data for all unconfirmed transactions from the mempool",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"unconfirmedtxs"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryPaginated{
					Page:    page,
					PerPage: per_page,
					Sort:    &sort,
				}

				response, err := client.PostV1QueryUnconfirmedtxs(cmd.Context(), body)
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
		{
			Use:     "Height",
			Short:   "Get current block height",
			Long:    "Queries the node RPC to obtain the current block height",
			Aliases: []string{"height"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}
				response, err := client.GetV1QueryHeight(cmd.Context())
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
		{
			Use:     "Transaction <hash>",
			Short:   "Get the transaction data the hash provided",
			Long:    "Queries the node RPC to obtain the transaction data for the specified hash",
			Aliases: []string{"tx"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryHash{
					Hash: args[0],
				}

				response, err := client.PostV1QueryTx(cmd.Context(), body)
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
			Use:     "UnconfirmedTransaction <hash>",
			Short:   "Get the unconfirmed transaction data the hash provided",
			Long:    "Queries the node RPC to obtain the unconfirmed transaction data for the specified hash, from the mempool",
			Aliases: []string{"unconfirmedtx"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.QueryHash{
					Hash: args[0],
				}

				response, err := client.PostV1QueryUnconfirmedtx(cmd.Context(), body)
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
