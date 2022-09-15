package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pokt-network/pocket/shared/test_artifacts"
)

// Utility to generate config and genesis files
// TODO(pocket/issues/182): Add a make target to help trigger this from cmdline

const (
	defaultGenesisFilePath = "build/config/genesis.json"
	defaultConfigFilePath  = "build/config/config"
	jsonSubfix             = ".json"
	rwoPerm                = 0777
)

func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(4, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(defaultGenesisFilePath, genesisJson, rwoPerm); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		filePath := fmt.Sprintf("%s%d%s", defaultConfigFilePath, i+1, jsonSubfix)
		if err := ioutil.WriteFile(filePath, configJson, rwoPerm); err != nil {
			panic(err)
		}
	}
}
