package cli

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/runtime/defaults"
)

const (
	cliExecutableName = "p1"
	flagBindErrFormat = "could not bind flag %q: %v"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flags.RemoteCLIURL, "remote_cli_url", defaults.DefaultRemoteCLIURL, "takes a remote endpoint in the form of <protocol>://<host>:<port> (uses RPC Port)")
	// ensure that this flag can be overridden by the respective viper-conventional environment variable (i.e. `POCKET_REMOTE_CLI_URL`)
	if err := viper.BindPFlag("remote_cli_url", rootCmd.PersistentFlags().Lookup("remote_cli_url")); err != nil {
		log.Fatalf(flagBindErrFormat, "remote_cli_url", err)
	}

	rootCmd.PersistentFlags().BoolVar(&flags.NonInteractive, "non_interactive", false, "if true skips the interactive prompts wherever possible (useful for scripting & automation)")

	// TECHDEBT: Why do we have a data dir when we have a config path if the data dir is only storing keys?
	rootCmd.PersistentFlags().StringVar(&flags.DataDir, "data_dir", defaults.DefaultRootDirectory, "Path to store pocket related data (keybase etc.)")
	rootCmd.PersistentFlags().StringVar(&flags.ConfigPath, "config", "", "Path to config")
	if err := viper.BindPFlag("root_directory", rootCmd.PersistentFlags().Lookup("data_dir")); err != nil {
		log.Fatalf(flagBindErrFormat, "data_dir", err)
	}

	rootCmd.PersistentFlags().BoolVar(&flags.Verbose, "verbose", false, "Show verbose output")
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		log.Fatalf(flagBindErrFormat, "verbose", err)
	}
}

var rootCmd = &cobra.Command{
	Use:               cliExecutableName,
	Short:             "Pocket Network Command Line Interface (CLI)",
	Long:              "The CLI is meant to be an user but also a machine friendly way for interacting with Pocket Network.",
	PersistentPreRunE: flags.ParseConfigAndFlags,
}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
