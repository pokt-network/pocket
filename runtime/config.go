package runtime

import (
	"encoding/json"
	"os"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	typesPers "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
	typesTelemetry "github.com/pokt-network/pocket/telemetry"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

var _ modules.ConsensusConfig = &Config{}
var _ modules.P2PConfig = &Config{}
var _ modules.PersistenceConfig = &Config{}
var _ modules.TelemetryConfig = &Config{}
var _ modules.UtilityConfig = &Config{}

type Config struct {
	Base        *BaseConfig                     `json:"base"`
	Consensus   *typesCons.ConsensusConfig      `json:"consensus"`
	Utility     *typesUtil.UtilityConfig        `json:"utility"`
	Persistence *typesPers.PersistenceConfig    `json:"persistence"`
	P2P         *typesP2P.P2PConfig             `json:"p2p"`
	Telemetry   *typesTelemetry.TelemetryConfig `json:"telemetry"`
}

func (c *Config) ToShared() modules.Config {
	return modules.Config{
		Base:        (*modules.BaseConfig)(c.Base),
		Consensus:   c.Consensus,
		Utility:     c.Utility,
		Persistence: c.Persistence,
		P2P:         c.P2P,
		Telemetry:   c.Telemetry,
	}
}

type BaseConfig struct {
	RootDirectory string `json:"root_directory"`
	PrivateKey    string `json:"private_key"` // TODO (pocket/issues/150) better architecture for key management (keybase, keyfiles, etc.)
	ConfigPath    string `json:"config_path"`
	GenesisPath   string `json:"genesis_path"`
}

func ParseConfigJSON(configPath string) (config *Config, err error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	// general configuration file
	config = new(Config)
	err = json.Unmarshal(data, &config)
	return
}

// modules.ConsensusConfig

func (c *Config) GetMaxMempoolBytes() uint64 {
	return c.Consensus.MaxMempoolBytes
}

// modules.P2PConfig

func (c *Config) GetConsensusPort() uint32 {
	return c.P2P.ConsensusPort
}

func (c *Config) IsEmptyConnType() bool { // TODO (team) make enum
	return c.P2P.IsEmptyConnectionType
}

// modules.PersistenceConfig

func (c *Config) GetPostgresUrl() string {
	return c.Persistence.PostgresUrl
}

func (c *Config) GetNodeSchema() string {
	return c.Persistence.NodeSchema
}

func (c *Config) GetBlockStorePath() string {
	return c.Persistence.BlockStorePath
}

// modules.TelemetryConfig

func (c *Config) GetEnabled() bool {
	return c.Telemetry.Enabled
}

func (c *Config) GetAddress() string {
	return c.Telemetry.Address
}

func (c *Config) GetEndpoint() string {
	return c.Telemetry.Endpoint
}
