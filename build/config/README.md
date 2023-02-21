# Genesis and Configuration Generator

The Genesis and Configuration generator creates V1 config and genesis files with coordinated key pairings for `localnet` and `devnet` usage.

The purpose of this utility is to enable efficient and accurate `auto-generation` of configuration/genesis files anytime the source structures change during V1 development.

## _Disclaimer_

Both the genesis and configuration contents and generation are living WIPs that are subject to rapid, breaking changes.

It is not recommended at this time to build infrastructure components that rely on the generator until it is stable

## Origin Document

Currently, the Genesis and Configuration generator is necessary to create development `localnet` environments for iterating on V1. A current example (as of 09/2022) of this is the `make compose_and_watch` debug utility that generates a `localnet` using `docker-compose` by injecting the appropriate `config.json` and `genesis.json` files.

## Usage

The output files are written to `./build/config/`.

### Using Source

From the project's root:

```bash
go run ./build/config/main.go --numFishermen=1
```

### Using Make Target

```bash
make numValidators=5 numServicers=1 gen_genesis_and_config
```

### Parameters

- `numValidators` is an int flag that sets the number of validators that will be in the network; this affects the contents of the genesis file as well as the number of config files
- `numServicers` is an int flag that set the number of servicers that will be in the network's genesis file
- `numApplications` is an int flag that set the number of applications that will be in the network's genesis file
- `numFishermen` is an int flag that set the number of fishermen that will be in the network's genesis file
- `genPrefix` is a string flag that adds a prefix to the generated files; is an empty string by default

## **WIP NOTE**

The config and genesis files located in the `./build/config/` directory are needed for following the local development instructions in `docs/development/README.md`.

These builds currently expect four (validator) `config.json` file and a single `genesis.json` file.

Until #186 is implemented, take caution when overwriting / deleting the files with different configurations.

<!-- GITHUB_WIKI: build/config/readme -->
