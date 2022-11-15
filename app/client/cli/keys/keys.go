package keys

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// keysCmd represents the base command when called without any subcommands
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Managing your public and private keys",
	Long:  `This is a key management CLI tool that supports multiple key management functions`,
}

func Execute() {
	err := keysCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent Flags
	keysCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.keys.yaml)")

	// Local flags
	keysCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Adding subcommands
	keysCmd.AddCommand(CreateCmd)
	keysCmd.AddCommand(DeleteCmd)
	keysCmd.AddCommand(MnemonicCmd)

	rootCmd.AddCommand(keysCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".keys" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".keys")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
