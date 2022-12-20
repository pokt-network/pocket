package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/pokt-network/pocket/internal/runtime/test_artifacts"
)

// Utility to generate config and genesis files

const (
	defaultGenesisFilePathFormat = "build/config/%sgenesis.json"
	defaultConfigFilePathFormat  = "build/config/%sconfig%d.json"
	rwoPerm                      = 0o777
)

var (
	numValidators   = flag.Int("numValidators", 4, "number of validators that will be in the network; this affects the contents of the genesis file as well as the # of config files")
	numServiceNodes = flag.Int("numServiceNodes", 1, "number of service nodes that will be in the network's genesis file")
	numApplications = flag.Int("numApplications", 1, "number of applications that will be in the network's genesis file")
	numFishermen    = flag.Int("numFishermen", 1, "number of fishermen that will be in the network's genesis file")
	genPrefix       = flag.String("genPrefix", "", "the prefix, if any, to append to the genesis and config files")
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
	if err = ioutil.WriteFile(fmt.Sprintf(defaultGenesisFilePathFormat, *genPrefix), genesisJson, rwoPerm); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		filePath := fmt.Sprintf(defaultConfigFilePathFormat, *genPrefix, i+1)
		if err := ioutil.WriteFile(filePath, configJson, rwoPerm); err != nil {
			panic(err)
		}
	}
}
