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
	DefaultGenesisFilePath = "build/config/genesis.json"
	DefaultConfigFilePath  = "build/config/config"
	JSONSubfix             = ".json"
	RWOPerm                = 0777
)

func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(4, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(DefaultGenesisFilePath, genesisJson, RWOPerm); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		filePath := fmt.Sprintf("%s%d%s", DefaultConfigFilePath, i+1, JSONSubfix)
		if err := ioutil.WriteFile(filePath, configJson, RWOPerm); err != nil {
			panic(err)
		}
	}
}
