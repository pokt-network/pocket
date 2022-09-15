## Genesis and Configuration Generator

The Genesis and Configuration generator creates V1 config and genesis files with coordinated key pairings for `localnet`
and `devnet` usage.

The purpose of this utility is to enable efficient and accurate `auto-generation` of configuration/genesis files anytime
the source structures change during V1 development.

### *Disclaimer*

Both the genesis and configuration contents and generation are living WIPs that are subject to rapid, breaking changes.

It is not recommended at this time to build infrastructure components that rely on the generator until it is stable

### Origin Document

Currently, the Genesis and Configuration generator is necessary to create development `localnet` environments
for implementing V1. A current example of this is the `make compose_and_watch` debug utility that generates a localnet
in a `docker compose` - injecting these appropriate config.json and genesis.json files

### Usage

From source at project root: `go run ./build/config/main.go --numFishermen=1`

From makefile at project root: `make numValidators=5 numServiceNodes=1 genesis_and_config`

The files output to the `./build/config/` directory

#### Parameters

`numValidators` is a string flag that sets the number of validators that will be in the network, this affects the
contents of the genesis file as well as the number of config files

`numServiceNodes` is a string flag that set the number of service nodes that will be in the network's genesis file

`numApplications` is a string flag that set the number of applications that will be in the network's genesis file

`numFishermen` is a string flag that set the number of fishermen that will be in the network's genesis file

### **NOTE**

The config and genesis files located in the `./build/config/` directory are needed for `make compose_and_watch` 
and `make client_start && make client_connect`. 

These builds expect four (valdiator) config and a single genesis file.

Take caution when overwriting / deleting the files with different configurations