package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/pokt-network/pocket/shared/test_artifacts"
)

// Utility to generate config and genesis files

const (
	defaultGenesisFilePath = "build/config/genesis.json"
	defaultConfigFilePath  = "build/config/config"
	jsonSuffix             = ".json"
	rwoPerm                = 0777
)

var (
	numValidators = flag.String("numValidators", "4", "number of validators that will be in the network, "+
		"this affects the contents of the genesis file as well as the # of config files")
	numServiceNodes = flag.String("numServiceNodes", "1", "number of service nodes that will be in the network's genesis file")
	numApplications = flag.String("numApplications", "1", "number of applications that will be in the network's genesis file")
	numFishermen    = flag.String("numFishermen", "1", "number of fishermen that will be in the network's genesis file")
)

func init() {
	flag.Parse()
}

func main() {
	nValidators, nServiceNodes, nFishermen, nApplications := catchEmptyFlags()
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(nValidators, nServiceNodes, nFishermen, nApplications)
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

func catchEmptyFlags() (nValidators, nServiceNodes, nFishermen, nApplications int) {
	if *numValidators == "" {
		*numValidators = "4"
	}
	if *numServiceNodes == "" {
		*numServiceNodes = "1"
	}
	if *numFishermen == "" {
		*numFishermen = "1"
	}
	if *numApplications == "" {
		*numApplications = "1"
	}
	nValidators, err := strconv.Atoi(*numValidators)
	if err != nil {
		log.Fatal("an error occurred when parsing number of validators: ", err.Error())
	}
	nServiceNodes, err = strconv.Atoi(*numServiceNodes)
	if err != nil {
		log.Fatal("an error occurred when parsing number of service nodes: ", err.Error())
	}
	nFishermen, err = strconv.Atoi(*numFishermen)
	if err != nil {
		log.Fatal("an error occurred when parsing number of fishermen: ", err.Error())
	}
	nApplications, err = strconv.Atoi(*numApplications)
	if err != nil {
		log.Fatal("an error occurred when parsing number of applications: ", err.Error())
	}
	return
}
