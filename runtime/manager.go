package runtime

import (
	"encoding/json"
	"io"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pokt-network/pocket/consensus/types"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/viper"
)

var _ modules.RuntimeMgr = &Manager{}

type Manager struct {
	config  *runtimeConfig
	genesis *runtimeGenesis
}

func NewManagerFromFiles(configPath, genesisPath string, options ...func(*Manager)) *Manager {
	rc := &Manager{}

	cfg, genesis, err := rc.init(configPath, genesisPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize runtime builder: %v", err)
	}
	rc.config = cfg
	rc.genesis = genesis

	for _, o := range options {
		o(rc)
	}

	return rc
}

func NewManagerFromReaders(configReader, genesisReader io.Reader, options ...func(*Manager)) *Manager {
	var cfg *runtimeConfig
	parse(configReader, cfg)

	var genesis *runtimeGenesis
	parse(genesisReader, genesis)

	rc := &Manager{
		config:  cfg,
		genesis: genesis,
	}
	return rc
}

func NewManager(config modules.Config, genesis modules.GenesisState) *Manager {
	return &Manager{
		config:  config.(*runtimeConfig),
		genesis: genesis.(*runtimeGenesis),
	}
}

func (rc *Manager) init(configPath, genesisPath string) (config *runtimeConfig, genesis *runtimeGenesis, err error) {
	dir, file := path.Split(configPath)
	filename := strings.TrimSuffix(file, filepath.Ext(file))

	viper.AddConfigPath(".")
	viper.AddConfigPath(dir)
	viper.SetConfigName(filename)
	viper.SetConfigType("json")

	// The lines below allow for environment variables configuration (12 factor app)
	// Eg: POCKET_CONSENSUS_PRIVATE_KEY=somekey would override `consensus.private_key` in config
	viper.SetEnvPrefix("POCKET")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	decoderConfig := func(dc *mapstructure.DecoderConfig) {
		// This is to leverage the `json` struct tags without having to add `mapstructure` ones.
		// Until we have complex use cases, this should work just fine.
		dc.TagName = "json"
	}
	if err = viper.Unmarshal(&config, decoderConfig); err != nil {
		return
	}

	if config.Base == nil {
		config.Base = &BaseConfig{}
	}

	genesis, err = parseGenesisJSON(genesisPath)
	return
}

func (b *Manager) GetConfig() modules.Config {
	return b.config
}

func (b *Manager) GetGenesis() modules.GenesisState {
	return b.genesis
}

type supportedStructs interface {
	*runtimeConfig | *runtimeGenesis
}

func parse[T supportedStructs](reader io.Reader, target T) {
	bz, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("[ERROR] Failed to read from reader: %v", err)

	}
	if err := json.Unmarshal(bz, &target); err != nil {
		log.Fatalf("[ERROR] Failed to unmarshal: %v", err)
	}
}

// Manager option helpers

func WithRandomPK() func(*Manager) {
	privateKey, err := cryptoPocket.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("unable to generate private key")
	}

	return WithPK(privateKey.String())
}

func WithPK(pk string) func(*Manager) {
	return func(b *Manager) {
		if b.config.Consensus == nil {
			b.config.Consensus = &types.ConsensusConfig{}
		}
		b.config.Consensus.PrivateKey = pk

		if b.config.P2P == nil {
			b.config.P2P = &typesP2P.P2PConfig{}
		}
		b.config.P2P.PrivateKey = pk
	}
}
