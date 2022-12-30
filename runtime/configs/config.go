package configs

import "github.com/pokt-network/pocket/runtime/defaults"

type Config struct {
	RootDirectory string `json:"root_directory"`
	PrivateKey    string `json:"private_key"` // INVESTIGATE(#150): better architecture for key management (keybase, keyfiles, etc.)

	Consensus   *ConsensusConfig   `json:"consensus"`
	Utility     *UtilityConfig     `json:"utility"`
	Persistence *PersistenceConfig `json:"persistence"`
	P2P         *P2PConfig         `json:"p2p"`
	Telemetry   *TelemetryConfig   `json:"telemetry"`
	Logger      *LoggerConfig      `json:"logger"`
	RPC         *RPCConfig         `json:"rpc"`
}

func NewDefaultConfig(options ...func(*Config)) *Config {
	cfg := &Config{
		RootDirectory: "/go/src/github.com/pocket-network",
		Consensus: &ConsensusConfig{
			MaxMempoolBytes: 500000000,
			PacemakerConfig: &PacemakerConfig{
				TimeoutMsec:               5000,
				Manual:                    true,
				DebugTimeBetweenStepsMsec: 1000,
			},
		},
		Utility: &UtilityConfig{
			MaxMempoolTransactionBytes: 1024 ^ 3, // 1GB V0 defaults
			MaxMempoolTransactions:     9000,
		},
		Persistence: &PersistenceConfig{
			PostgresUrl:    "postgres://postgres:postgres@pocket-db:5432/postgres",
			BlockStorePath: "/var/blockstore",
		},
		P2P: &P2PConfig{
			ConsensusPort:         8080,
			UseRainTree:           true,
			IsEmptyConnectionType: false,
			MaxMempoolCount:       defaults.DefaultP2PMaxMempoolCount,
		},
		Telemetry: &TelemetryConfig{
			Enabled:  true,
			Address:  "0.0.0.0:9000",
			Endpoint: "/metrics",
		},
		Logger: &LoggerConfig{
			Level:  "debug",
			Format: "pretty",
		},
		RPC: &RPCConfig{
			Timeout: defaults.DefaultRpcTimeout,
			Port:    defaults.DefaultRpcPort,
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
