package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"strconv"

	"github.com/pokt-network/pocket/shared/test_artifacts"
)

// Utility to generate config and genesis files

const (
	defaultGenesisFilePath = "build/config/gen.genesis.json"
	defaultConfigFilePath  = "build/config/gen.config"
	jsonSuffix             = ".json"
	rwoPerm                = 0777
)

var (
	numValidators   = flag.Int("numValidators", 4, "number of validators that will be in the network; this affects the contents of the genesis file as well as the # of config files")
	numServiceNodes = flag.Int("numServiceNodes", 1, "number of service nodes that will be in the network's genesis file")
	numApplications = flag.Int("numApplications", 1, "number of applications that will be in the network's genesis file")
	numFishermen    = flag.Int("numFishermen", 1, "number of fishermen that will be in the network's genesis file")
)

func init() {
	flag.Parse()
}

func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(*numValidators, *numServiceNodes, *numFishermen, *numApplications)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(defaultGenesisFilePath, genesisJson, rwoPerm); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(defaultConfigFilePath+strconv.Itoa(i+1)+jsonSuffix, configJson, rwoPerm); err != nil {
			panic(err)
		}
	}
}
