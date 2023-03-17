package configs

import (
	"log"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/spf13/viper"
)

// IMPROVE: add a SaveConfig() function to save the config to a file and
// generate a default config file for the user. Add it as a new command
// into the CLI.

type Config struct {
	RootDirectory   string `json:"root_directory"`
	PrivateKey      string `json:"private_key"` // INVESTIGATE(#150): better architecture for key management (keybase, keyfiles, etc.)
	ClientDebugMode bool   `json:"client_debug_mode"`
	UseLibP2P       bool   `json:"use_lib_p2p"` // Determines if `root/libp2p` or `root/p2p` should be used as the p2p module

	Consensus   *ConsensusConfig   `json:"consensus"`
	Utility     *UtilityConfig     `json:"utility"`
	Persistence *PersistenceConfig `json:"persistence"`
	P2P         *P2PConfig         `json:"p2p"`
	Telemetry   *TelemetryConfig   `json:"telemetry"`
	Logger      *LoggerConfig      `json:"logger"`
	RPC         *RPCConfig         `json:"rpc"`
	Keybase     *KeybaseConfig     `json:"keybase"`
}

// ParseConfig parses the config file and returns a Config struct
func ParseConfig(cfgFile string) *Config {
	config := NewDefaultConfig()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {

		viper.AddConfigPath("/etc/pocket/")  // path to look for the config file in
		viper.AddConfigPath("$HOME/.pocket") // call multiple times to add many search paths
		viper.AddConfigPath(".")             // optionally look for config in the working directory
		viper.SetConfigName("config")        // name of config file (without extension)
		viper.SetConfigType("json")          // REQUIRED if the config file does not have the extension in the name
	}

	// The lines below allow for environment variables configuration (12 factor app)
	// Eg: POCKET_CONSENSUS_PRIVATE_KEY=somekey would override `consensus.private_key` in config
	viper.SetEnvPrefix("POCKET")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && cfgFile == "" {
			// if file does not exist and there is no override, return default config
			// DISCUSS: how should we handle this? Should we return an error? Should we create the default file?
			return config
		} else {
			// Some other error occurred while reading the config file
			log.Fatalf("[ERROR] failed to read config %s", err.Error())
		}
	}

	decoderConfig := func(dc *mapstructure.DecoderConfig) {
		// This is to leverage the `json` struct tags without having to add `mapstructure` ones.
		// Until we have complex use cases, this should work just fine.
		dc.TagName = "json"
	}
	if err := viper.Unmarshal(&config, decoderConfig); err != nil {
		log.Fatalf("[ERROR] failed to unmarshal config %s", err.Error())
	}
	return config
}

// IMPROVE: could go all in on viper and use viper.SetDefault()
func NewDefaultConfig(options ...func(*Config)) *Config {
	cfg := &Config{
		RootDirectory: defaults.DefaultRootDirectory,
		UseLibP2P:     defaults.DefaultUseLibp2p,
		Consensus: &ConsensusConfig{
			MaxMempoolBytes: defaults.DefaultConsensusMaxMempoolBytes,
			PacemakerConfig: &PacemakerConfig{
				TimeoutMsec:               defaults.DefaultPacemakerTimeoutMsec,
				Manual:                    defaults.DefaultPacemakerManual,
				DebugTimeBetweenStepsMsec: defaults.DefaultPacemakerDebugTimeBetweenStepsMsec,
			},
		},
		Utility: &UtilityConfig{
			MaxMempoolTransactionBytes: defaults.DefaultUtilityMaxMempoolTransactionBytes,
			MaxMempoolTransactions:     defaults.DefaultUtilityMaxMempoolTransactions,
		},
		Persistence: &PersistenceConfig{
			PostgresUrl:    defaults.DefaultPersistencePostgresURL,
			BlockStorePath: defaults.DefaultPersistenceBlockStorePath,
		},
		P2P: &P2PConfig{
			Port:            defaults.DefaultP2PPort,
			UseRainTree:     defaults.DefaultP2PUseRainTree,
			ConnectionType:  defaults.DefaultP2PConnectionType,
			MaxMempoolCount: defaults.DefaultP2PMaxMempoolCount,
		},
		Telemetry: &TelemetryConfig{
			Enabled:  defaults.DefaultTelemetryEnabled,
			Address:  defaults.DefaultTelemetryAddress,
			Endpoint: defaults.DefaultTelemetryEndpoint,
		},
		Logger: &LoggerConfig{
			Level:  defaults.DefaultLoggerLevel,
			Format: defaults.DefaultLoggerFormat,
		},
		RPC: &RPCConfig{
			Timeout: defaults.DefaultRPCTimeout,
			Port:    defaults.DefaultRPCPort,
		},
		Keybase: &KeybaseConfig{
			KeybaseType:    defaults.DefaultKeybaseType,
			KeybasePath:    defaults.DefaultKeybasePath,
			VaultAddr:      defaults.DefaultKeybaseVaultAddr,
			VaultToken:     defaults.DefaultKeybaseVaultToken,
			VaultMountPath: defaults.DefaultKeybaseVaultMountPath,
		},
	}

	for _, option := range options {
		option(cfg)
	}

	return cfg
}

// WithPK is an option to configure module-specific keys in order to enable different "identities"
// for different purposes (i.e. validation, P2P, servicing, etc...).
func WithPK(pk string) func(*Config) {
	return func(cfg *Config) {
		cfg.PrivateKey = pk
		cfg.Consensus.PrivateKey = pk
		cfg.P2P.PrivateKey = pk
	}
}

// WithNodeSchema is an option to configure the schema for the node's database.
func WithNodeSchema(schema string) func(*Config) {
	return func(cfg *Config) {
		cfg.Persistence.NodeSchema = schema
	}
}
