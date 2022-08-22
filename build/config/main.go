package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
)

const (
	filePermissions    = 0777
	genesisPath        = "build/config/genesis.json"
	configPathTemplate = "build/config/config%d.json"
)

// Utility to generate config and genesis files
// TODO_IN_THIS_COMMIT: Add a make target to help trigger this from cmdline
func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(4, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(genesisPath, genesisJson, filePermissions); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(fmt.Sprintf(configPathTemplate, i+1), configJson, filePermissions); err != nil {
			panic(err)
		}
	}
}
