package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	pcrypto "github.com/pokt-network/pocket/shared/crypto"
)

type Config struct {
	RootDir string `json:"root_dir"`
	Genesis string `json:"genesis"`

	PrivateKey pcrypto.Ed25519PrivateKey `json:"private_key"`

	Pre2P       *Pre2PConfig       `json:"pre2p"` // TODO(derrandz): delete this once P2P is ready.
	P2P         *P2PConfig         `json:"p2p"`
	Consensus   *ConsensusConfig   `json:"consensus"`
	Persistence *PersistenceConfig `json:"persistence"`
	Utility     *UtilityConfig     `json:"utility"`
}

// TODO(derrandz): delete this once P2P is ready.
type Pre2PConfig struct {
	ConsensusPort uint32 `json:"consensus_port"`
	DebugPort     uint32 `json:"debug_port"`
}

type P2PConfig struct {
	Protocol   string          `json:"protocol"`
	Address    pcrypto.Address `json:"address"`
	ExternalIp string          `json:"external_ip"`
	Peers      []string        `json:"peers"`
}

type PacemakerConfig struct {
	TimeoutMsec               uint64 `json:"timeout_msec"`
	Manual                    bool   `json:"manual"`
	DebugTimeBetweenStepsMsec uint64 `json:"debug_time_between_steps_msec"`
}

type ConsensusConfig struct {
	Pacemaker *PacemakerConfig `json:"pacemaker"`
}

type PersistenceConfig struct {
	DataDir string `json:"datadir"`
}

type UtilityConfig struct {
}

// TODO(insert tooling issue # here): Re-evaluate how load configs should be handeled.
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

	if err := c.validateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing config: ", err)
	}

	if err := c.Consensus.validateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing consensus config: ", err)
	}

	if err := c.P2P.validateAndHydrate(); err != nil {
		log.Fatalln("Error validating or completing P2P config: ", err)
	}

	return
}

func (c *Config) validateAndHydrate() error {
	if len(c.PrivateKey) == 0 {
		return fmt.Errorf("private key in config file cannot be empty")
	}

	if len(c.Genesis) == 0 {
		return fmt.Errorf("must specify a genesis file")
	}
	c.Genesis = rootify(c.Genesis, c.RootDir)

	return nil
}

func (c *P2PConfig) validateAndHydrate() error {
	return nil
}

func (c *ConsensusConfig) validateAndHydrate() error {
	if err := c.Pacemaker.validateAndHydrate(); err != nil {
		log.Fatalf("Error validating or completing Pacemaker configs")
	}
	return nil
}

func (c *PacemakerConfig) validateAndHydrate() error {
	return nil
}

// Helper to make config creation independent of root dir
func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
