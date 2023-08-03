package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
)

// Utility to generate config and genesis files

const (
	defaultGenesisFilePathFormat = "build/config/%sgenesis.json"
	defaultConfigFilePathFormat  = "build/config/%sconfig%d.json"
	rwoPerm                      = 0o0777
)

var (
	numValidators   = flag.Int("numValidators", 4, "number of validators that will be in the network; this affects the contents of the genesis file as well as the # of config files")
	numServicers    = flag.Int("numServicers", 1, "number of servicers that will be in the network's genesis file")
	numApplications = flag.Int("numApplications", 1, "number of applications that will be in the network's genesis file")
	numWatchers     = flag.Int("numWatchers", 1, "number of watchers that will be in the network's genesis file")
	genPrefix       = flag.String("genPrefix", "", "the prefix, if any, to append to the genesis and config files")
)

func init() {
	flag.Parse()
}

func main() {
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(*numValidators, *numServicers, *numWatchers, *numApplications)
	configs := test_artifacts.NewDefaultConfigs(validatorPrivateKeys)
	genesisJson, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(fmt.Sprintf(defaultGenesisFilePathFormat, *genPrefix), genesisJson, rwoPerm); err != nil {
		panic(err)
	}
	for i, config := range configs {
		configJson, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		filePath := fmt.Sprintf(defaultConfigFilePathFormat, *genPrefix, i+1)
		if err := os.WriteFile(filePath, configJson, rwoPerm); err != nil {
			panic(err)
		}
	}
}
