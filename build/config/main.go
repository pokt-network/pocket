package main

import (
	"encoding/json"
	"flag"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"io/ioutil"
	"log"
	"strconv"
)

// Utility to generate config and genesis files

const (
	DefaultGenesisFilePath = "build/config/genesis.json"
	DefaultConfigFilePath  = "build/config/config"
	JSONSubfix             = ".json"
	RWOPerm                = 0777
)

var (
	numValidators = flag.String("numValidators", "4", "set the number of validators that will be in the network, "+
		"this affects the contents of the genesis file as well as the # of config files")
	numServiceNodes = flag.String("numServiceNodes", "1", "set the number of service nodes that will be in the network's genesis file")
	numApplications = flag.String("numApplications", "1", "set the number of applications that will be in the network's genesis file")
	numFishermen    = flag.String("numFishermen", "1", "set the number of fishermen that will be in the network's genesis file")
)

func init() {
	flag.Parse()
}

func main() {
	nValidators, nServiceNodes, nFishermen, nApplications := catchEmptyFlags()
	genesis, validatorPrivateKeys := test_artifacts.NewGenesisState(nValidators, nServiceNodes, nFishermen, nApplications)
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
