package main

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"io/ioutil"
	"strconv"
)

// Utility to generate config and genesis files
// TODO(andrew): Add a make target to help trigger this from cmdline

const (
	DefaultGenesisFilePath = "build/config/genesis.json"
	DefaultConfigFilePath  = "build/config/config"
	JSONSubfix             = ".json"
	RWOPerm                = 0777
)

func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(4, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, _ := json.MarshalIndent(genesis, "", "  ")
	if err := ioutil.WriteFile(DefaultGenesisFilePath, genesisJson, RWOPerm); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, _ := json.MarshalIndent(config, "", "  ")
		if err := ioutil.WriteFile(DefaultConfigFilePath+strconv.Itoa(i+1)+JSONSubfix, configJson, RWOPerm); err != nil {
			panic(err)
		}
	}
}
