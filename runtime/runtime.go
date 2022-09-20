package runtime

import (
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/spf13/viper"
)

var _ modules.Runtime = &Builder{}

type Builder struct {
	configPath  string
	genesisPath string

	config  *Config
	genesis *Genesis

	useRandomPK bool
}

func NewBuilder(configPath, genesisPath string, options ...func(*Builder)) *Builder {
	b := &Builder{
		configPath:  configPath,
		genesisPath: genesisPath,
	}

	cfg, genesis, err := b.init()
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize runtime builder: %v", err)
	}
	b.config = cfg
	b.genesis = genesis

	for _, o := range options {
		o(b)
	}

	return b
}

func (b *Builder) init() (config *Config, genesis *Genesis, err error) {
	dir, file := path.Split(b.configPath)
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

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config, func(dc *mapstructure.DecoderConfig) {
		// This is to leverage the `json` struct tags without having to add `mapstructure` ones.
		// Until we have complex use cases, this should work just fine.
		dc.TagName = "json"
	})
	if err != nil {
		return
	}

	if config.Base == nil {
		config.Base = &BaseConfig{}
	}
	config.Base.ConfigPath = b.configPath
	config.Base.GenesisPath = b.genesisPath

	genesis, err = ParseGenesisJSON(b.genesisPath)
	return
}

func (b *Builder) GetConfig() modules.Config {
	return b.config.ToShared()
}

func (b *Builder) GetGenesis() modules.GenesisState {
	return b.genesis.ToShared()
}

func (b *Builder) ShouldUseRandomPK() bool {
	return b.useRandomPK
}

func WithRandomPK() func(*Builder) {
	return func(b *Builder) { b.useRandomPK = true }
}
