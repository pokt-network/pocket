package runtime

func Init(configPath, genesisPath string) (config *Config, genesis *Genesis, err error) {
	config, err = ParseConfigJSON(configPath)
	if err != nil {
		return
	}

	if config.Base == nil {
		config.Base = &BaseConfig{}
	}
	config.Base.ConfigPath = configPath
	config.Base.GenesisPath = genesisPath

	genesis, err = ParseGenesisJSON(genesisPath)
	return
}
