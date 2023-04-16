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
	chain         string
	geozone       string
	sessionHeight int64
)

func init() {
	queryCmd := NewClientCommand()
	rootCmd.AddCommand(queryCmd)
}

func NewClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Client",
		Short:   "Commands related to sending requests to the network via the node's RPC server",
		Aliases: []string{"client"},
		Args:    cobra.ExactArgs(0),
	}

	dispatchCmds := clientDispatchCommand()

	// attach --chain flag
	applySubcommandOptions(dispatchCmds, attachChainFlagToSubcommands())

	// attach --geozone flag
	applySubcommandOptions(dispatchCmds, attachGeoZoneFlagToSubcommands())

	// attach --session_height flag
	applySubcommandOptions(dispatchCmds, attachSessionHeightFlagToSubcommands())

	cmd.AddCommand(dispatchCmds...)

	return cmd
}

func clientDispatchCommand() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Dispatch <address> [--chain] [--geozone] [--session_height]",
			Short:   "Send a dispatch request to the network",
			Long:    "Sends a dispatch request to the node's RPC server and returns session data",
			Args:    cobra.ExactArgs(1),
			Aliases: []string{"dispatch"},
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}

				body := rpc.DispatchRequest{
					AppAddress:    args[0],
					Chain:         chain,
					Geozone:       geozone,
					SessionHeight: sessionHeight,
				}

				response, err := client.PostV1ClientDispatch(cmd.Context(), body)
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
