package cli

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/pocket/rpc"
	"github.com/spf13/cobra"
)

func init() {
	systemCmd := NewSystemCommand()
	rootCmd.AddCommand(systemCmd)
}

func NewSystemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "System",
		Short:   "Commands related to health and troubleshooting of the node instance",
		Aliases: []string{"sys"},
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(systemCommands()...)

	return cmd
}

func systemCommands() []*cobra.Command {
	cmds := []*cobra.Command{
		{
			Use:     "Health",
			Short:   "RPC endpoint liveness",
			Long:    "Performs a simple liveness check on the node RPC endpoint",
			Aliases: []string{"health"},
			RunE: func(cmd *cobra.Command, args []string) error {

				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return nil
				}
				response, err := client.GetV1HealthWithResponse(cmd.Context())
				if err != nil {
					return unableToConnectToRpc(err)
				}
				statusCode := response.StatusCode()
				if statusCode == http.StatusOK {
					fmt.Printf("✅ RPC reporting healthy status for node @ %s\n\n%s", boldText(remoteCLIURL), response.Body)
					return nil
				}

				return rpcResponseCodeUnhealthy(statusCode, response.Body)
			},
		},
		{
			Use:     "Version",
			Short:   "Advertised node software version",
			Long:    "Queries the node RPC to obtain the version of the software currently running",
			Aliases: []string{"version"},
			RunE: func(cmd *cobra.Command, args []string) error {

				client, err := rpc.NewClientWithResponses(remoteCLIURL)
				if err != nil {
					return err
				}
				response, err := client.GetV1VersionWithResponse(cmd.Context())
				if err != nil {
					return unableToConnectToRpc(err)
				}
				statusCode := response.StatusCode()
				if statusCode == http.StatusOK {
					fmt.Printf("Node @ %s reports that it's running version: \n%s\n", boldText(remoteCLIURL), boldText(response.Body))
					return nil
				}

				return rpcResponseCodeUnhealthy(statusCode, response.Body)
			},
		},
	}
	return cmds
}
