package runtime

func Init(configPath, genesisPath string) (config *Config, genesis *Genesis, err error) {
	config, err = ParseConfigJSON(configPath)
	if err != nil {
		return
	}
	config.configPath = configPath
	config.genesisPath = genesisPath

	genesis, err = ParseGenesisJSON(genesisPath)
	return
}
