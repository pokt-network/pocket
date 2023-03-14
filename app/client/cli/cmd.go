package cli

import (
	"context"
	"os"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/spf13/cobra"
)

const (
	cliExecutableName = "client"
	keybaseSuffix     = "/keys"
)

var (
	remoteCLIURL   string
	dataDir        string
	nonInteractive bool
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("Cannot find user home directory")
	}
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remote_cli_url", defaults.DefaultRemoteCLIURL, "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non_interactive", false, "if true skips the interactive prompts wherever possible (useful for scripting & automation)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data_dir", homeDir+"/.pocket", "Path to store pocket related data (keybase etc.)")
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
