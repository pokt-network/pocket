package cli

import (
	"context"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliExecutableName = "client"
)

var (
	remoteCLIURL   string
	dataDir        string
	configPath     string
	nonInteractive bool
	cfg            *configs.Config
)

func init() {
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remote_cli_url", defaults.DefaultRemoteCLIURL, "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non_interactive", false, "if true skips the interactive prompts wherever possible (useful for scripting & automation)")

	// TECHDEBT: Why do we have a data dir when we have a config path if the data dir is only storing keys?
	rootCmd.PersistentFlags().StringVar(&dataDir, "data_dir", defaults.DefaultRootDirectory, "Path to store pocket related data (keybase etc.)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config")
	if err := viper.BindPFlag("root_directory", rootCmd.PersistentFlags().Lookup("data_dir")); err != nil {
		panic(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   cliExecutableName,
	Short: "Pocket Network Command Line Interface (CLI)",
	Long:  "The CLI is meant to be an user but also a machine friendly way for interacting with Pocket Network.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// by this time, the config path should be set
		cfg = configs.ParseConfig(configPath)
		return nil
	},
}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
