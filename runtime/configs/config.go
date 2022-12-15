package configs

type Config struct {
	RootDirectory string `json:"root_directory"`
	PrivateKey    string `json:"private_key"` // TODO (#150) better architecture for key management (keybase, keyfiles, etc.)

	Consensus   *ConsensusConfig   `json:"consensus"`
	Utility     *UtilityConfig     `json:"utility"`
	Persistence *PersistenceConfig `json:"persistence"`
	P2P         *P2PConfig         `json:"p2p"`
	Telemetry   *TelemetryConfig   `json:"telemetry"`
	Logger      *LoggerConfig      `json:"logger"`
	RPC         *RPCConfig         `json:"rpc"`
}
