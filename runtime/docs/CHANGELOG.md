# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.40] - 2023-05-31

- Add a new Address field to Servicer configuration
- Add a place-holder default servicer configuration

## [0.0.0.39] - 2023-05-25

- Added a new ServicerConfig type

## [0.0.0.38] - 2023-05-08

- Renamed `P2PConfig#MaxMempoolCount` to `P2PConfig#MaxNonces`
- Renamed `DefaultP2PMaxMempoolCount` to `DefaultP2PMaxNonces`

## [0.0.0.37] - 2023-05-04

- Add `network_id` field to the configs and give the default value of `"localnet"`

## [0.0.0.36] - 2023-04-28

- Consolidated files for defaults together
- Updated `BlocksPerSession` default to 1
- Added an ability to add options to the `NewGenesisState` helper for more thorough testing

## [0.0.0.35] - 2023-04-27

- Removed unneeded `use_rain_tree` P2P config field

## [0.0.0.34] - 2023-04-19

- Changed `Validator1EndpointK8S` which now reflects the new value.

## [0.0.0.33] - 2023-04-17

- Removed `runtime/configs.Config#UseLibp2p` field

## [0.0.0.32] - 2023-04-13

- Add persistent txIndexerPaths to node configs and update tests

## [0.0.0.31] - 2023-04-11

- Add comment regarding KeybaseConfig proto design.

## [0.0.0.30] - 2023-04-07

- Update `genesis.proto` to add `owner` tags to all governance parameters

## [0.0.0.29] - 2023-04-06

- Updated to reflect pools address changes

## [0.0.0.28] - 2023-03-30

- Update the configurations for postgres pooling

## [0.0.0.27] - 2023-03-28

- Adds keybase_config.proto

## [0.0.0.26] - 2023-03-26

- Updated defaults to be `const` instead of `var` where applicable

## [0.0.0.25] - 2023-03-01

- replace `consensus_port` with `port` in P2P config
- update default P2P config `port` to from `8080` to `42069`
- add `use_libp2p` field to base config
- add `hostname` field to P2P config

## [0.0.0.24] - 2023-02-28

- Rename `app_staking_adjustment` to `app_session_tokens_multiplier`
- Remove `app_baseline_stake_rate`
- Rename `keygenerator` to `keygen`

## [0.0.0.23] - 2023-02-21

- Rename ServiceNode Actor Type Name to Servicer

## [0.0.0.22] - 2023-02-20

- Added `bootstrap_nodes_csv` in `P2PConfig` to allow for a comma separated list of bootstrap nodes

## [0.0.0.21] - 2023-02-17

- Added validator accounts from the genesis file to the `manager_test.go`

## [0.0.0.20] - 2023-02-17

- Nits: variables visibility, comments

## [0.0.0.19] - 2023-02-17

- Introduced `modules.ModulesRegistry` for better separation of concerns
- Added `StateMachineModule` accessors
- `Manager` embeds `base_modules.IntegratableModule` for DRYness

## [0.0.0.18] - 2023-02-16

- Added `IsProcessRunningInsideKubernetes` and centralized `GetEnv` so that it can be used across the board

## [0.0.0.17] - 2023-02-14

- Move shared utils (e.g. `BitIngToString`) to the `converters` package
- Remove `CleanupTest`

## [0.0.0.16] - 2023-02-09

- Update runtime consensus config with bool server mode variable
- Update manager test

## [0.0.0.15] - 2023-02-07

- Added GITHUB_WIKI tags where it was missing

## [0.0.0.14] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.13] - 2023-02-06

- Added additional logging information to be able to tell which config file contains an error
- Changed hardcoded addresses and public keys to reflect new addresses pattern from LocalNet on Kubernetes

## [0.0.0.12] - 2023-02-04

- Changed log lines to utilize new logger module.

## [0.0.0.11] - 2023-02-03

- Updated to display the warning message about the telemetry module not registered only once

## [0.0.0.10] - 2023-01-25

- move ConnectionType enum into its own package to avoid a cyclic import between configs and defaults packages (i.e. configs -> defaults -> configs) in the resulting, generated go package
- update makefile protogen_local target to build additional proto file and include it in the import path for runtime/configs/proto/p2p_config.proto
- replace `P2PConfig#IsEmptyConnectionType` bool with `P2PConfig#ConnectionType` enum
- replace `DefaultP2PIsEmptyConnectionType` bool with `DefaultP2PConnectionType` enum

## [0.0.0.9] - 2023-01-23

- Updated README.md with information about node profiling

## [0.0.0.8] - 2023-01-19

- Rewrite `interface{}` to `any`

## [0.0.0.7] - 2023-01-14

- Added MaxConnsCount, MinConnsCount, MaxConnLifetime, MaxConnIdleTime, and HealthCheckPeriod to persistence config.

## [0.0.0.6] - 2023-01-11

- Updated tests to reflect the updated genesis file

## [0.0.0.5] - 2023-01-10

- Updated modules constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering modules with the `bus` via `RegisterModule` method
- Providing Dependency Injection functionality via `bus`
- Updated tests and mocks accordingly

## [0.0.0.4] - 2023-01-09

- Added 'is_client_only' to `P2PConfig`

## [0.0.0.3] - 2023-01-03

- Split testing/development configs into separate files
- Centralized `NewDefaultConfig` logic with options used by the config generator
- Refactored Params handling, not hardcoded anymore but sourced from genesis

## [0.0.0.2] - 2022-12-21

- Centralized config handling into a `config` package
- Config protos from the various modules are now in the `config` package
- Removed the `BaseConfig` struct
- Removed overlapping parts in `PersistenceGenesisState` and `ConsensusGenesisState` and consolidated under a single `GenesisState` struct
- Updated tests to use the new config and genesis handling
- Introduced a singleton `keyGenerator` capable of generating keys randomly or deterministically (#414)

## [0.0.0.1] - 2022-12-14

- Added `DefaultP2PMaxMempoolCount`

## [0.0.0.0] - 2022-09-30

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Abstracted config and genesis handling
- Mockable runtime
- Refactored all modules to use `RuntimeMgr`
- Updated `RuntimeMgr` to manage clock as well
- Modules now accept `interfaces` instead of paths.
- Unmarshalling is done in a new `runtime` package (runtime because what we do in there affects the runtime of the application)
- We are now able to accept configuration via environment variables (thanks to @okdas for inspiration and [sp13 for Viper]("github.com/spf13/viper"))

<!-- GITHUB_WIKI: changelog/runtime -->
