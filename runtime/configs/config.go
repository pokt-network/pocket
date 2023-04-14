package configs

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/mitchellh/mapstructure"
	// "github.com/pokt-network/pocket/logger"
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

	Consensus   *ConsensusConfig   `json:"consensus"`
	Utility     *UtilityConfig     `json:"utility"`
	Persistence *PersistenceConfig `json:"persistence"`
	P2P         *P2PConfig         `json:"p2p"`
	Telemetry   *TelemetryConfig   `json:"telemetry"`
	Logger      *LoggerConfig      `json:"logger"`
	RPC         *RPCConfig         `json:"rpc"`
	Keybase     *KeybaseConfig     `json:"keybase"` // Determines and configures which keybase to use, `file` or `vault`. IMPROVE(#626): See for rationale around proto design. We have proposed a better config design, but did not implement it due to viper limitations
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
	// Eg: POCKET_CONSENSUS_PRIVATE_KEY=somekey would override `consensus.private_key` in config.json
	// If the key is not set in the config, the env var will not be used.
	viper.SetEnvPrefix("POCKET")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && cfgFile == "" {
			log.Default().Printf("No config provided, using defaults")
		} else {
			// TODO: This is a log call to avoid import cycles. Refactor logger_config.proto to avoid this.
			log.Fatalf("[ERROR] fatal error reading config file %s", err.Error())
		}
	} else {
		// TODO: This is a log call to avoid import cycles. Refactor logger_config.proto to avoid this.
		log.Default().Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	decoderConfig := func(dc *mapstructure.DecoderConfig) {
		// This is to leverage the `json` struct tags without having to add `mapstructure` ones.
		// Until we have complex use cases, this should work just fine.
		dc.TagName = "json"
	}
	// Detect if we need to use json.Unmarshal instead of viper.Unmarshal
	if err := viper.Unmarshal(&config, decoderConfig); err != nil {
		cfgData := viper.AllSettings()
		cfgJSON, _ := json.Marshal(cfgData)

		// last ditch effort to unmarshal the config
		if err := json.Unmarshal(cfgJSON, &config); err != nil {
			log.Fatalf("[ERROR] failed to unmarshal config %s", err.Error())
		}
	}

	return config
}

// setViperDefaults this is a hacky way to set the default values for Viper so env var overrides work.
// DISCUSS: is there a better way to do this?
func setViperDefaults(cfg *Config) {
	// convert the config struct to a map with the json tags as keys
	cfgData, err := json.Marshal(cfg)

	if err != nil {
		log.Fatalf("[ERROR] failed to marshal config %s", err.Error())
	}
	var cfgMap map[string]any
	if err := json.Unmarshal(cfgData, &cfgMap); err != nil {
		log.Fatalf("[ERROR] failed to unmarshal config %s", err.Error())
	}

	for k, v := range cfgMap {
		viper.SetDefault(k, v)
	}
}

func NewDefaultConfig(options ...func(*Config)) *Config {
	cfg := &Config{
		RootDirectory: defaults.DefaultRootDirectory,
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
			Type:           defaults.DefaultKeybaseType,
			FilePath:       defaults.DefaultKeybaseFilePath,
			VaultAddr:      defaults.DefaultKeybaseVaultAddr,
			VaultToken:     defaults.DefaultKeybaseVaultToken,
			VaultMountPath: defaults.DefaultKeybaseVaultMountPath,
		},
	}

	for _, option := range options {
		option(cfg)
	}

	// set Viper defaults so POCKET_ env vars work without having to set in config file
	setViperDefaults(cfg)

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
