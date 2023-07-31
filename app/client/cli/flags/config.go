package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pokt-network/pocket/runtime/configs"
)

var Cfg *configs.Config

func ParseConfigAndFlags(_ *cobra.Command, _ []string) error {
	// by this time, the config path should be set
	Cfg = configs.ParseConfig(ConfigPath)

	// set final `remote_cli_url` value; order of precedence: flag > env var > config > default
	RemoteCLIURL = viper.GetString("remote_cli_url")
	return nil
}
