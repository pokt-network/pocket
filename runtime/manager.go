package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/mitchellh/mapstructure"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/viper"
)

var _ modules.RuntimeMgr = &Manager{}

type Manager struct {
	modules.BaseIntegratableModule

	config       *configs.Config
	genesisState *genesis.GenesisState

	clock clock.Clock
	bus   modules.Bus
}

func NewManager(config *configs.Config, genesis *genesis.GenesisState, options ...func(*Manager)) *Manager {
	mgr := new(Manager)
	bus, err := CreateBus(mgr)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize bus: %v", err)
	}

	mgr.config = config
	mgr.genesisState = genesis
	mgr.clock = clock.New()
	mgr.bus = bus

	for _, o := range options {
		o(mgr)
	}

	return mgr
}

func NewManagerFromFiles(configPath, genesisPath string, options ...func(*Manager)) *Manager {
	cfg, genesisState, err := parseFiles(configPath, genesisPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize runtime builder: %v", err)
	}
	return NewManager(cfg, genesisState, options...)
}

// NewManagerFromReaders returns a *Manager given io.Readers for the config and the genesis.
//
// Useful for testing and when the user doesn't want to rely on the filesystem and instead intends plugging in different configuration management system.
func NewManagerFromReaders(configReader, genesisReader io.Reader, options ...func(*Manager)) *Manager {
	cfg := configs.NewDefaultConfig()
	parseFromReader(configReader, cfg)

	genesisState := new(genesis.GenesisState)
	parseFromReader(genesisReader, genesisState)

	return NewManager(cfg, genesisState, options...)
}

func (m *Manager) GetConfig() *configs.Config {
	return m.config
}

func (m *Manager) GetGenesis() *genesis.GenesisState {
	return m.genesisState
}

func (m *Manager) GetClock() clock.Clock {
	return m.clock
}

func parseFiles(configJSONPath, genesisJSONPath string) (config *configs.Config, genesisState *genesis.GenesisState, err error) {
	config = configs.NewDefaultConfig()

	dir, configFile := path.Split(configJSONPath)
	filename := strings.TrimSuffix(configFile, filepath.Ext(configFile))

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
		err = fmt.Errorf("error reading %s: %w", configJSONPath, err)
		return
	}

	decoderConfig := func(dc *mapstructure.DecoderConfig) {
		// This is to leverage the `json` struct tags without having to add `mapstructure` ones.
		// Until we have complex use cases, this should work just fine.
		dc.TagName = "json"
	}
	if err = viper.Unmarshal(&config, decoderConfig); err != nil {
		err = fmt.Errorf("error unmarshalling %s: %w", configJSONPath, err)
		return
	}

	genesisState, err = parseGenesis(genesisJSONPath)
	return
}

func parseFromReader[T *configs.Config | *genesis.GenesisState](reader io.Reader, target T) {
	bz, err := io.ReadAll(reader)
	if err != nil {
		logger.Global.Err(err).Msg("Failed to read from reader")
	}
	if err := json.Unmarshal(bz, target); err != nil {
		logger.Global.Err(err).Msg("Failed to unmarshal")
	}
}

// Manager option helpers

func WithRandomPK() func(*Manager) {
	privateKey, err := cryptoPocket.GeneratePrivateKey()
	if err != nil {
		logger.Global.Fatal().Err(err).Msg("unable to generate private key")
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

func WithClientDebugMode() func(*Manager) {
	return func(b *Manager) {
		b.config.ClientDebugMode = true
	}
}
