package configs

import "github.com/pokt-network/pocket/runtime/defaults"

type Config struct {
	RootDirectory   string `json:"root_directory"`
	PrivateKey      string `json:"private_key"` // INVESTIGATE(#150): better architecture for key management (keybase, keyfiles, etc.)
	ClientDebugMode bool   `json:"client_debug_mode"`
	UseLibP2P       bool   `json:"use_lib_p2p"` // Determines if `root/libp2p` or `root/p2p` should be used as the p2p module

	Consensus     *ConsensusConfig   `json:"consensus"`
	Utility       *UtilityConfig     `json:"utility"`
	Persistence   *PersistenceConfig `json:"persistence"`
	P2P           *P2PConfig         `json:"p2p"`
	Telemetry     *TelemetryConfig   `json:"telemetry"`
	Logger        *LoggerConfig      `json:"logger"`
	RPC           *RPCConfig         `json:"rpc"`
	KeybaseConfig *KeybaseConfig     `json:"keybase_config"`
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
