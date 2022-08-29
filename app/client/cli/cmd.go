package cli

import (
	"context"

	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	"github.com/spf13/cobra"
)

const cliExecutableName = "client"

var (
	remoteCLIURL       string
	privateKeyFilePath string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remoteCLIURL", test_artifacts.DefaultRemoteCliUrl, "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")
	rootCmd.PersistentFlags().StringVar(&privateKeyFilePath, "path_to_private_key_file", "./pk.json", "Path to private key to use when signing")
}

var rootCmd = &cobra.Command{
	Use: cliExecutableName,
	// TODO(deblasis): document
	Short: "",
	Long:  "",
	// TODO(deblasis): do we need some sort of teardown as well?

}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
