package cli

import (
	"context"

	"github.com/pokt-network/pocket/app"
	"github.com/spf13/cobra"
)

const CLIExecutableName = "client"

var (
	remoteCLIURL string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remoteCLIURL", "", "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")

	rootCmd.AddCommand(NewAccountCommand())
	rootCmd.AddCommand(NewActorCommands()...)
	rootCmd.AddCommand(NewGovernanceCommand())
	rootCmd.AddCommand(NewDebug())
}

var rootCmd = &cobra.Command{
	Use: CLIExecutableName,
	// TODO(deblasis): document
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		app.Init(remoteCLIURL)
	},
	// TODO(deblasis): do we need some sort of teardown as well?

}

func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
