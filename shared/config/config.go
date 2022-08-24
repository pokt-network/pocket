package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pokt-network/pocket/shared/crypto"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

type Config struct {
	RootDir       string                 `json:"root_dir"`
	GenesisSource *genesis.GenesisSource `json:"genesis_source"` // TECHDEBT(olshansky): we should be able to pass the struct in here.

	RPCConfig *RPCConfig `json:"rpc"`

	EnableTelemetry bool                           `json:"enable_telemetry"`
	PrivateKey      cryptoPocket.Ed25519PrivateKey `json:"private_key"`

	P2P         *P2PConfig         `json:"p2p"`
	Consensus   *ConsensusConfig   `json:"consensus"`
	Persistence *PersistenceConfig `json:"persistence"`
	Utility     *UtilityConfig     `json:"utility"`
	Telemetry   *TelemetryConfig   `json:"telemetry"`
}

type ConnectionType string

const (
	TCPConnection   ConnectionType = "tcp"
	EmptyConnection ConnectionType = "empty" // Only used for testing
)

// TECHDEBT(team): consolidate/replace this with P2P configs depending on next steps
type P2PConfig struct {
	ConsensusPort  uint32         `json:"consensus_port"`
	UseRainTree    bool           `json:"use_raintree"`
	ConnectionType ConnectionType `json:"connection_type"`
}

type PacemakerConfig struct {
	TimeoutMsec               uint64 `json:"timeout_msec"`
	Manual                    bool   `json:"manual"`
	DebugTimeBetweenStepsMsec uint64 `json:"debug_time_between_steps_msec"`
}

type ConsensusConfig struct {
	// Mempool
	MaxMempoolBytes uint64 `json:"max_mempool_bytes"` // TODO(olshansky): add unit tests for this

	// Block
	MaxBlockBytes uint64 `json:"max_block_bytes"` // TODO(olshansky): add unit tests for this

	// Pacemaker
	Pacemaker *PacemakerConfig `json:"pacemaker"`
}

type PersistenceConfig struct {
	PostgresUrl    string `json:"postgres_url"`
	NodeSchema     string `json:"schema"`
	BlockStorePath string `json:"block_store_path"`
}

type UtilityConfig struct {
}

type TelemetryConfig struct {
	Address  string // The address the telemetry module will use to listen for metrics PULL requests (e.g. 0.0.0.0:9000 for prometheus)
	Endpoint string // The endpoint available to fetch recorded metrics (e.g. /metrics for prometheus)
}

type RPCConfig struct {
	Enable       bool   `json:"enable"`
	Port         string `json:"port"`           // TODO(deblasis): still unused, need to define default value, "8081" ?
	RemoteCLIURL string `json:"remote_cli_url"` // TODO(deblasis): need to define default value as const somewhere, "http://localhost:8081" ?
	Timeout      int64  `json:"timeout"`        // TODO(deblasis): need to define default value as const somewhere, "30000" ?
}

// TODO(insert tooling issue # here): Re-evaluate how load configs should be handeled.
func LoadConfig(file string) (c *Config) {
	// Not used - only set to avoid `GetNodeState(_)` from crashing
	pk, err := crypto.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("[ERROR] Failed to generate private key: %v", err)
	}

	c = New(
		WithPrivateKey(pk.(crypto.Ed25519PrivateKey)),
	)

	jsonFile, err := os.Open(file)
	if err != nil {
		log.Fatalln("Error opening config file: ", err)
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalln("Error reading config file: ", err)
	}
	if err = json.Unmarshal(bytes, c); err != nil {
		log.Fatalln("Error parsing config file: ", err)
	}

	if err := c.ValidateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing config: ", err)
	}

	return
}

// TODO: Exhaust all the configuration validation checks
func (c *Config) ValidateAndHydrate() error {
	if len(c.PrivateKey) == 0 {
		return fmt.Errorf("private key in config file cannot be empty")
	}

	if c.GenesisSource == nil {
		return fmt.Errorf("genesis source cannot be nil in config")
	}

	if err := c.HydrateGenesisState(); err != nil {
		return fmt.Errorf("error getting genesis state: %v", err)
	}

	if err := c.Consensus.ValidateAndHydrate(); err != nil {
		return fmt.Errorf("error validating or completing consensus config: %v", err)
	}

	return nil
}

func (c *Config) HydrateGenesisState() error {
	genesisState, err := genesis.GenesisStateFromGenesisSource(c.GenesisSource)
	if err != nil {
		return fmt.Errorf("error getting genesis state: %v", err)
	}
	c.GenesisSource.Source = &genesis.GenesisSource_State{
		State: genesisState,
	}
	return nil
}

func (c *ConsensusConfig) ValidateAndHydrate() error {
	if err := c.Pacemaker.ValidateAndHydrate(); err != nil {
		log.Fatalf("Error validating or completing Pacemaker configs")
	}

	if c.MaxMempoolBytes <= 0 {
		return fmt.Errorf("MaxMempoolBytes must be a positive integer")
	}

	if c.MaxBlockBytes <= 0 {
		return fmt.Errorf("MaxBlockBytes must be a positive integer")
	}

	return nil
}

func (c *PacemakerConfig) ValidateAndHydrate() error {
	return nil
}

// New simply returns the initialized struct with default values
func New(options ...func(*Config)) *Config {
	c := &Config{
		GenesisSource: &genesis.GenesisSource{
			Source: &genesis.GenesisSource_File{
				File: &genesis.GenesisFile{
					Path: "build/config/genesis.json",
				},
			},
		},

		// Used to access the validator map
		Consensus: &ConsensusConfig{
			Pacemaker: &PacemakerConfig{},
		},

		// Not used - only set to avoid `p2p.Create()` from crashing
		P2P: &P2PConfig{
			ConsensusPort:  9999,
			UseRainTree:    true,
			ConnectionType: TCPConnection,
		},
		EnableTelemetry: false,

		RPCConfig: &RPCConfig{
			Port:         defaultRPCPort,
			RemoteCLIURL: defaultRemoteCLIURL,
			Timeout:      defaultRPCTimeout,
		},
	}
	for _, o := range options {
		o(c)
	}
	return c
}

func WithPrivateKey(pk crypto.PrivateKey) func(c *Config) {
	return func(c *Config) {
		// Not used - only set to avoid `GetNodeState(_)` from crashing
		c.PrivateKey = pk.(crypto.Ed25519PrivateKey)
	}
}
