package cli

import (
	"context"

	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remote_cli_url", defaults.DefaultRemoteCLIURL, "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non_interactive", false, "if true skips the interactive prompts wherever possible (useful for scripting & automation)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data_dir", defaults.DefaultRootDirectory, "Path to store pocket related data (keybase etc.)")
	viper.BindPFlag("root_directory", rootCmd.PersistentFlags().Lookup("data_dir"))
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
