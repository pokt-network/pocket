package cli

import (
	"context"

	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/spf13/cobra"
)

const cliExecutableName = "client"

var (
	remoteCLIURL       string
	privateKeyFilePath string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remote_cli_url", defaults.DefaultRemoteCLIURL, "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")
	rootCmd.PersistentFlags().StringVar(&privateKeyFilePath, "path_to_private_key_file", "./pk.json", "Path to private key to use when signing")
}

var rootCmd = &cobra.Command{
	Use:   cliExecutableName,
	Short: "Pocket Network Command Line Interface (CLI)",
	Long:  "The CLI is meant to be an user but also a machine friendly way for interacting with Pocket Network.",
}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
