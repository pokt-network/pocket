package main

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"io/ioutil"
	"strconv"
)

// Utility to generate config and genesis files
func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(4, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, _ := json.MarshalIndent(genesis, "", "  ")
	if err := ioutil.WriteFile("build/config/genesis.json", genesisJson, 0777); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, _ := json.MarshalIndent(config, "", "  ")
		if err := ioutil.WriteFile("build/config/config"+strconv.Itoa(i+1)+".json", configJson, 0777); err != nil {
			panic(err)
		}
	}
}
