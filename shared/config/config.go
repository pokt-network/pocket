package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

type Config struct {
	RootDir       string                 `json:"root_dir"`
	GenesisSource *genesis.GenesisSource `json:"genesis_source"` // TECHDEBT(olshansky): we should be able to pass the struct in here.

	PrivateKey cryptoPocket.Ed25519PrivateKey `json:"private_key"`

	// TECHDEBT(team): Consolidate `Pre2P` and `P2P`
	Pre2P     *Pre2PConfig     `json:"pre2p"`
	P2P       *P2PConfig       `json:"p2p"`
	Consensus *ConsensusConfig `json:"consensus"`
	// TECHDEBT(team): Consolidate `Persistence` and `PrePersistence`
	PrePersistence *PrePersistenceConfig `json:"pre_persistence"`
	Persistence    *PersistenceConfig    `json:"persistence"`
	Utility        *UtilityConfig        `json:"utility"`
}

type ConnectionType string

const (
	TCPConnection   ConnectionType = "tcp"
	EmptyConnection ConnectionType = "empty" // Only used for testing
)

// TECHDEBT(team): consolidate/replace this with P2P configs depending on next steps
type Pre2PConfig struct {
	ConsensusPort  uint32         `json:"consensus_port"`
	UseRainTree    bool           `json:"use_raintree"`
	ConnectionType ConnectionType `json:"connection_type"`
}

type PrePersistenceConfig struct {
	Capacity        int `json:"capacity"`
	MempoolMaxBytes int `json:"mempool_max_bytes"`
	MempoolMaxTxs   int `json:"mempool_max_txs"`
}

type P2PConfig struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	// TODO(derrandz): Fix the config imports appropriately
	// Address          cryptoPocket.Address `json:"address"`
	ExternalIp       string   `json:"external_ip"`
	Peers            []string `json:"peers"`
	MaxInbound       uint32   `json:"max_inbound"`
	MaxOutbound      uint32   `json:"max_outbound"`
	BufferSize       uint     `json:"connection_buffer_size"`
	WireHeaderLength uint     `json:"max_wire_header_length"`
	TimeoutInMs      uint     `json:"timeout_in_ms"`
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
	DataDir     string `json:"datadir"`
	PostgresUrl string `json:"postgres_url"`
	NodeSchema  string `json:"schema"`
}

type UtilityConfig struct {
}

func LoadConfig(file string) (c *Config) {
	c = &Config{}

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
		return fmt.Errorf("error hydrating genesis state: %v", err)
	}

	if err := c.Consensus.ValidateAndHydrate(); err != nil {
		return fmt.Errorf("error validating or completing consensus config: %v", err)
	}

	if err := c.P2P.ValidateAndHydrate(); err != nil {
		return fmt.Errorf("error validating or completing P2P config: %v", err)
	}

	return nil
}

func (c *Config) HydrateGenesisState() error {
	var genesisState *genesis.GenesisState
	var err error

	switch c.GenesisSource.Source.(type) {
	case *genesis.GenesisSource_Config:
		genesisConfig := c.GenesisSource.GetConfig()
		if genesisState, _, _, _, _, err = genesis.NewGenesisState(genesisConfig); err != nil {
			return fmt.Errorf("failed to generate genesis state from configuration: %v", err)
		}
	case *genesis.GenesisSource_File:
		genesisFilePath := c.GenesisSource.GetFile().Path
		if _, err := os.Stat(genesisFilePath); err != nil {
			return fmt.Errorf("genesis file specified but not found %s", genesisFilePath)
		}
		if genesisState, err = genesis.GenesisStateFromFile(genesisFilePath); err != nil {
			return fmt.Errorf("failed to load genesis state from file: %v", err)
		}
	case *genesis.GenesisSource_State:
		genesisState = c.GenesisSource.GetState()
	default:
		return fmt.Errorf("unsupported genesis source type: %v", c.GenesisSource.Source)
	}

	c.GenesisSource.Source = &genesis.GenesisSource_State{
		State: genesisState,
	}
	return nil
}

func (c *P2PConfig) ValidateAndHydrate() error {
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
