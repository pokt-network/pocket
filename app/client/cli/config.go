package cli

import (
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(saveDefaultConf)
}

var saveDefaultConf = &cobra.Command{
	Use:   "save_default_config",
	Short: "Save the default config in a file",
	Long:  "The default config generated during application start is saved in a config file path passed in the argument",
	Run: func(cmd *cobra.Command, args []string) {
		configs.SaveConfig(args[0])
	},
}
