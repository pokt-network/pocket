package runtime

import (
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/viper"
)

var _ modules.Runtime = &RuntimeConfig{}

type RuntimeConfig struct {
	configPath  string
	genesisPath string

	config  *Config
	genesis *Genesis

	useRandomPK bool
}

func New(configPath, genesisPath string, options ...func(*RuntimeConfig)) *RuntimeConfig {
	rc := &RuntimeConfig{
		configPath:  configPath,
		genesisPath: genesisPath,
	}

	cfg, genesis, err := rc.init()
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

func (rc *RuntimeConfig) init() (config *Config, genesis *Genesis, err error) {
	dir, file := path.Split(rc.configPath)
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
	config.Base.ConfigPath = rc.configPath
	config.Base.GenesisPath = rc.genesisPath

	genesis, err = ParseGenesisJSON(rc.genesisPath)
	return
}

func (b *RuntimeConfig) GetConfig() modules.Config {
	return b.config.ToShared()
}

func (b *RuntimeConfig) GetGenesis() modules.GenesisState {
	return b.genesis.ToShared()
}

func (b *RuntimeConfig) ShouldUseRandomPK() bool {
	return b.useRandomPK
}

func WithRandomPK() func(*RuntimeConfig) {
	return func(b *RuntimeConfig) { b.useRandomPK = true }
}

func WithPK(pk string) func(*RuntimeConfig) {
	return func(b *RuntimeConfig) {
		if b.config.Consensus == nil {
			b.config.Consensus = &types.ConsensusConfig{}
		}
		b.config.Consensus.PrivateKey = pk
	}
}
