package configs

import (
	"encoding/json"
	"fmt"
	"log"
	reflect "reflect"
	"strings"
	"unicode"

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
			Config: &KeybaseConfig_File{
				File: &KeybaseFileConfig{
					Path: defaults.DefaultKeybaseFilePath,
				},
			},
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

// decoderConfig allows for complete control over the decoding process.
func decoderConfig(dc *mapstructure.DecoderConfig) {
	// This is to leverage the `json` struct tags without having to add `mapstructure` ones.
	// Until we have complex use cases, this should work just fine.
	dc.TagName = "json"

	// Add decode hook for KeybaseConfig
	dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		dc.DecodeHook,
		decodeKeybaseConfig,
	)

	// This is to allow for case-insensitive and snake-case matching of config keys.
	dc.MatchName = customMatchFunc
}

// decodeKeybaseConfig is a custom decode hook for KeybaseConfig due to the fact KeybaseConfig contains a generic(oneof) field.
func decodeKeybaseConfig(f, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(KeybaseConfig{}) {
		return data, nil
	}

	m := data.(map[string]interface{})
	keybaseConfig := KeybaseConfig{}

	if config, ok := m["config"]; ok {
		configMap := config.(map[string]interface{})

		// check vault first since it is not the default
		if vaultConfig, ok := configMap["vault"]; ok {
			keybaseVaultConfig := &KeybaseVaultConfig{}
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				MatchName: customMatchFunc,
				Result:    keybaseVaultConfig,
			})
			if err != nil {
				return nil, err
			}
			if err := decoder.Decode(vaultConfig); err != nil {
				return nil, err
			}
			keybaseConfig.Config = &KeybaseConfig_Vault{Vault: keybaseVaultConfig}
			return &keybaseConfig, nil // return early
		}

		// other keybase config types should go here before the default

		// check file **last** since it is the default
		if fileConfig, ok := configMap["file"]; ok {
			keybaseFileConfig := &KeybaseFileConfig{}
			if err := mapstructure.Decode(fileConfig, keybaseFileConfig); err != nil {
				return nil, err
			}
			keybaseConfig.Config = &KeybaseConfig_File{File: keybaseFileConfig}
			return &keybaseConfig, nil // return early
		}
	}

	// should never happen, this means the default keybase config logic was changed
	return nil, fmt.Errorf("unsupported keybase config type")
}

// Custom match function to handle both mount_path and mountPath
func customMatchFunc(mapKey, fieldName string) bool {
	if strings.EqualFold(mapKey, fieldName) {
		return true
	}

	// Convert field name from CamelCase to snake_case
	snakeFieldName := convertCamelToSnake(fieldName)

	// Check if the snake_case version of the field name matches the map key
	return strings.EqualFold(mapKey, snakeFieldName)
}

// Convert CamelCase string to snake_case string
func convertCamelToSnake(input string) string {
	var sb strings.Builder

	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 {
				sb.WriteRune('_')
			}
			sb.WriteRune(unicode.ToLower(r))
		} else {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}
