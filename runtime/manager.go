package runtime

import (
	"encoding/json"
	"io"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/mitchellh/mapstructure"
	"github.com/pokt-network/pocket/runtime/configs"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/viper"
)

var _ modules.RuntimeMgr = &Manager{}

type Manager struct {
	config  *configs.Config
	genesis *runtimeGenesis

	clock clock.Clock
}

func NewManagerFromFiles(configPath, genesisPath string, options ...func(*Manager)) *Manager {
	mgr := &Manager{
		clock: clock.New(),
	}

	cfg, genesis, err := mgr.init(configPath, genesisPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize runtime builder: %v", err)
	}
	mgr.config = cfg
	mgr.genesis = genesis

	for _, o := range options {
		o(mgr)
	}

	return mgr
}

// NewManagerFromReaders returns a *Manager given io.Readers for the config and the genesis.
//
// Ideally useful when the user doesn't want to rely on the filesystem and instead intends plugging in different configuration management system.
//
// Note: currently unused, here as a reference
func NewManagerFromReaders(configReader, genesisReader io.Reader, options ...func(*Manager)) *Manager {
	var cfg *configs.Config
	parse(configReader, cfg)

	var genesis *runtimeGenesis
	parse(genesisReader, genesis)

	mgr := &Manager{
		config:  cfg,
		genesis: genesis,
		clock:   clock.New(),
	}

	for _, o := range options {
		o(mgr)
	}

	return mgr
}

func NewManager(config *configs.Config, genesis modules.GenesisState, options ...func(*Manager)) *Manager {
	mgr := &Manager{
		config:  config,
		genesis: genesis.(*runtimeGenesis),
		clock:   clock.New(),
	}

	for _, o := range options {
		o(mgr)
	}

	return mgr
}

func (rc *Manager) init(configPath, genesisPath string) (config *configs.Config, genesis *runtimeGenesis, err error) {
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

	genesis, err = parseGenesisJSON(genesisPath)
	return
}

func (b *Manager) GetConfig() *configs.Config {
	return b.config
}

func (b *Manager) GetGenesis() modules.GenesisState {
	return b.genesis
}

func (b *Manager) GetClock() clock.Clock {
	return b.clock
}

type supportedStructs interface {
	*configs.Config | *runtimeGenesis
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
			b.config.Consensus = &configs.ConsensusConfig{}
		}
		b.config.Consensus.PrivateKey = pk

		if b.config.P2P == nil {
			b.config.P2P = &configs.P2PConfig{}
		}
		b.config.P2P.PrivateKey = pk
	}
}

func WithClock(clockMgr clock.Clock) func(*Manager) {
	return func(b *Manager) {
		b.clock = clockMgr
	}
}
