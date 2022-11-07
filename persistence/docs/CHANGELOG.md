# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**TODO:** consolidate `persistence/docs/CHANGELOG` and `persistence/CHANGELOG.md`

## [Unreleased]

## [0.0.0.9] - 2022-11-04

- Changed the following exposed functions to lowercase non-exported functions
- [./pocket/persistence/]
  - GetActor
  - GetActorFromRow
  - GetChainsForActor
  - SetActorStakeAmount
  - GetActorStakeAmount
  - GetCtxAndTx
  - GetCtx
  - SetValidatorStakedTokens
  - GetValidatorStakedTokens
- [./pocket/persistence/types]
  - ProtocolActorTableSchema
  - ProtocolActorChainsTableSchema
  - SelectChains
  - ReadyToUnstake
  - InsertChains
  - UpdateUnstakingHeight
  - UpdateStakeAmount
  - UpdatePausedHeight
  - UpdateUnstakedHeightIfPausedBefore
  - AccToAccInterface
  - TestInsertParams
  - AccountOrPoolSchema
  - InsertAcc
  - SelectBalance
- [./pocket/persistence/test]
  - GetGenericActor
  - NewTestGenericActor

## [0.0.0.8] - 2022-10-19

- Fixed `ToPersistenceActors()` by filling all structure fields
- Deprecated `BaseActor` -> `Actor`
- Changed default actor type to `ActorType_Undefined`

## [0.0.0.7] - 2022-10-12

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`
- Made `PersistenceModule` struct unexported
- Updated tests and mocks
- Removed some cross-module dependencies
- Added `TxIndexer` sub-package (previously in Utility Module)
- Added `TxIndexer` to both `PersistenceModule` and `PersistenceContext`
- Implemented `TransactionExists` and `StoreTransaction`

## [0.0.0.6] - 2022-09-30

- Removed no-op `DeleteActor` code
- Consolidated `CHANGELOG`s into one under `persistence/docs`
- Consolidated `README`s into one under `persistence/docs`
- Deprecated `persMod.ResetContext()` for -> `persRWContext.ResetContext()` for more appropriate encapsulation
- Added ticks to CHANGELOG.md
- Removed reference to Utility Mod's `BigIntToString()` and used internal `BigIntToString()`

## [0.0.0.5] - 2022-08-25

**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**

- Renamed schema -> types
- Added genesis, config, and unstaking proto files from shared
- Ensured proto structures implement shared interfaces
- Populate `PersistenceGenesisState` uses shared interfaces in order to accept `MockPersistenceGenesisState`
- ^ Same applies for `PersistenceConfig`
- Bumped cleanup TODOs to #149 due to scope size of #163

## [0.0.0.4] - 2022-08-16

**Main persistence module changes:**

- Split `ConnectAndInitializeDatabase` into `connectToDatabase` and `initializeDatabase`
    - This enables creating multiple contexts in parallel without re-initializing the DB connection
- Fix the SQL query used in `SelectActors`, `SelectAccounts` & `SelectPools`
    - Add a generalized unit test for all actors
- Remove `NewPersistenceModule` and an injected `Config` + `Create`
    - This improves isolation a a “injection-like” paradigm for unit testing
- Change `SetupPostgresDocker` to `SetupPostgresDockerPersistenceMod`
    - This enables more “functional” like testing by returning a persistence module and avoiding global testing
      variables
    - Only return once a connection to the DB has been initialized reducing the likelihood of test race conditions
- Implemented `NewReadContext` with a proper read-only context
- Add `ResetContext` to the persistence module and `Close` to the read context

**Secondary persistence module changes**

- Improve return values in `Commit` and `Release` (return error, add logging, etc…)
- Add `pgx.Conn` pointer to `PostgresDB`
- `s/db/conn/g` and `s/conn/tx/g` in some (not all) places where appropriate
- Make some exported variables / functions unexported for readability & access purposes
- Add a few helpers for persistence related unit testing
- Added unit tests and TODOs for handling multiple read/write contexts

## [0.0.0.3] - 2022-08-08

- Deprecated old placeholder `genesis_state` and `genesis_config`
- Added `utility_genesis_state` to `genesis_state`
- Added `consensus_genesis_state` to `genesis_state`
- Added `genesis_time` to `consensus_genesis_state`
- Added `chainID` to `consensus_genesis_state`
- Added `max_block_bytes` to `consensus_genesis_state`
- Added `accounts` and` pools to utility_genesis_state`
- Added `validators` to `utility_genesis_state`
- Added `applications` to `utility_genesis_state`
- Added `service_nodes` to `utility_genesis_state`
- Added `fishermen` to `utility_genesis_state`
- Deprecated `shared/config/`
- Added new `shared config proto3 structure`
- Added `base_config` to `config`
- Added `utility_config` to `config`
- Added `consensus_config` to `config`
- Added `persistence_config` to `config`
- Added `p2p_config` to `config`
- Added `telemetry_config` to `config`
- Opened followup issue #163
- Added config and genesis generator to build package
- Deprecated old build files
- Use new config and genesis files for make `compose_and_watch`
- Use new config and genesis files for make `client_start && `make client_connect`

## [0.0.0.2] - 2022-08-03

Deprecate PrePersistence

- Fix for bytes parameters
- Accounts / pools default to 0
- Pre-added accounts to genesis file
- Separated out Persistence Read Context from Persistence Write Context
- Added various TODO's in order to code-complete a working persistence module
- Added genesis level functions to GetAllActors() and GetAllAccounts/Pools() for testing
- Added PopulateGenesisState function to persistence module
- Fixed the stake status iota issue
- Discovered and documented (with TODO) double setting parameters issue
- Attached to the Utility Module and using in `make compose_and_watch`

## [0.0.0.1] - 2022-07-05

Pocket Persistence 1st Iteration (https://github.com/pokt-network/pocket/pull/73)

- Base persistence module implementation for the following actors: `Account`, `Pool`, `Validator`, `Fisherman`
  , `ServiceNode`, `Application`
- Generalization of common protocol actor behvaiours via the `ProtocolActor` and `BaseActor` interface and
  implementation
- A PostgreSQL based implementation of the persistence middleware including:
    - SQL query implementation for each actor
    - SQL schema definition for each actor
    - SQL execution for common actor behaviours
    - Golang interface implementation of the Persistence module
- Update to the Persistence module interface to enable historical height queries
- Library / infrastructure for persistence unit fuzz testing
- Tests triggered via `make test_persistence`
