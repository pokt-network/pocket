# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.36] - 2023-02-17

- Module now embeds `base_modules.IntegratableModule` for DRYness

## [0.0.0.35] - 2023-02-15

- Add a few `nolint` comments to fix the code on main

## [0.0.0.34] - 2023-02-14

- Remove `IUnstakingActor` and use `UnstakingActor` directly; guideline for removing future unnecessary types (e.g. TxResult)
- Typo in `GetMinimumBlockHeightQuery`
- Reduce unnecessary `string` <-> `[]byte` conversion in a few places
- Fix bug in `updateUnstakedHeightIfPausedBefore` that was unstaking all actors

## [0.0.0.33] - 2023-02-09

- Added mock generation to the `kvstore/kvstore.go`.

## [0.0.0.32] - 2023-02-07

- Minor documentation cleanup

## [0.0.0.31] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.30] - 2023-02-04

- Changed log lines to utilize new logger module

## [0.0.0.29] - 2023-01-31

- Use hash of serialised protobufs for keys in `updateParamsTree()` and `updateFlagsTree()`

## [0.0.0.28] - 2023-01-30

- Fix unit tests - `TestGetAppPauseHeightIfExists`, `TestGetAppOutputAddress`, `TestGetFishermanStatus`, `TestGetFishermanPauseHeightIfExists`, `TestGetFishermanOutputAddress`, `TestPersistenceContextParallelReadWrite`, `TestGetServicerPauseHeightIfExists`, `TestGetServicerOutputAddress`, `fuzzSingleProtocolActor`, `TestGetValidatorPauseHeightIfExists`, and `TestGetValidatorOutputAddress` for misplaced expected and actual values in `require.Equal`.

## [0.0.0.27] - 2023-01-27

- Add logic for `updateParamsTree()` and `updateFlagsTree()` functions when updating merkle root hash

## [0.0.0.26] - 2023-01-23

- Added `debug.FreeOSMemory()` on `ResetToGenesis` to free-up memory and stabilize `LocalNet`.

## [0.0.0.25] - 2023-01-20

- Consolidate common behaviour of `Pool` and `Account` functions into a shared interface `ProtocolAccountSchema`
- Create `account_shared_sql.go` and `types/account_shared_sql.go` and rename `shared_sql.go` and `type/shared_sql.go` to `actor_shared_sql.go` and `types/actor_shared_sql.go` seperating shared sql logic

## [0.0.0.24] - 2023-01-20

- Update the persistence module README, focusing on `pgadmin` and Makefile helpers

## [0.0.0.23] - 2023-01-18

- Remove `Block` proto definition to consolidate under `shared/core/types`

## [0.0.0.22] - 2023-01-14

- Add `max_conns_count`, `min_conns_count`, `max_conn_lifetime`, `max_conn_idle_time` and `health_check_period` to `PersistenceConfig`.
- Update `connectToDatabase` function in `db.go` to connect via `pgxpool` to postgres database and accept `PersistenceConfig` interface as input.
- Update `github.com/jackc/pgx/v4` -> `github.com/jackc/pgx/v5`.

## [0.0.0.21] - 2023-01-11

- Add `init()` function to `gov.go` to build a map of parameter names and their types
- Deprecated `GetBlocksPerSession()` and `GetServicersPerSessionAt()` in favour of the more general parameter getter function `GetParameter()`
- Update unit tests replacing `GetIntParam()` and `GetStringParam()` calls with `GetParameter()`

## [0.0.0.20] - 2023-01-11

- Minor logging improvements

## [0.0.0.19] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering module with the `bus` via `RegisterModule` method
- Updated tests and mocks accordingly

## [0.0.0.18] - 2023-01-03

- Renamed `InitParams` to `InitGenesisParams`

## [0.0.0.17] - 2023-01-03

- Added missing `ActorType` in `GetAllXXXX()` functions
- Updated to new `PoolNames` enums
- Using `Enum.FriendlyName()` instead of `Enum.String()` for `PoolNames` enums (backward compatibility + flexibility)
- Updated `InitParams` so that Params can be initialized from a `GenesisState` and not just hardcoded
- Refactored default values sourcing (test_artifacts for tests)
- Updated tests
- Consolidated `persistence/docs/CHANGELOG` and `persistence/CHANGELOG.md` into `persistence/docs/CHANGELOG`

## [0.0.0.16] - 2022-12-21

- Updated to use centralized config and genesis
- Updated to use `Account` struct now under `coreTypes`
- Tended for the TODO "// TODO (Andrew) genericize the genesis population logic for actors #149" in `persistence/genesis.go`
- Updated tests to use the new config and genesis handling
- Updated statetest hashes to reflect updated genesis state

## [0.0.0.15] - 2022-12-15

- Remove `SetProposalBlock` and local vars to keep proposal state
- Add `proposerAddr` to the `Commit` function
- Move the `PostgresContext` struct to `context.db`

## [0.0.0.14] - 2022-12-14

- Moved Actor related getters from `genesis.go` to `actor.go`
- Added `GetAllStakedActors()` that returns all Actors

## [0.0.0.13] - 2022-12-06

- Changed the scope of `TransactionExists` from the `PostgresContext` to the `PersistenceModule`

## [0.0.0.13] - 2022-11-30

Core StateHash changes

- Introduced & defined for `block_persistence.proto`
  - A persistence specific protobuf for the Block stored in the BlockStore
- On `Commit`, prepare and store a persistence block in the KV Store, SQL Store
- Replace `IndexTransactions` (plural) to `IndexTransaction` (singular)
- Maintaining a list of StateTrees using Celestia’s SMT and badger as the KV store to compute the state hash
- Implemented `ComputeStateHash` to update the global state based on:
  - Validators
  - Applications
  - Servicers
  - Fisherman
  - Accounts
  - Pools
  - Transactions
  - Added a placeholder for `params` and `flags`
- Added a benchmarking and a determinism test suite to validate this

Supporting StateHash changes

- Implemented `GetAccountsUpdated`, `GetPoolsUpdated` and `GetActorsUpdated` functions
- Removed `GetPrevAppHash` and `indexTransactions` functions
- Removed `blockProtoBytes` and `txResults` from the local state and added `quorumCert`
- Consolidate all `resetContext` related operations into a single function
- Implemented `ReleaseWriteContext`
- Implemented ability to `ClearAllState` and `ResetToGenesis` for debugging & testing purposes
- Added unit tests for all of the supporting SQL functions implemented
- Some improvements in unit test preparation & cleanup (limited to this PR's functionality)

KVStore changes

- Renamed `Put` to `Set`
- Embedded `smt.MapStore` in the interface containing `Get`, `Set` and `Delete`
- Implemented `Delete`
- Modified `GetAll` to return both `keys` and `values`
- Turned off badger logging options since it’s noisy

## [0.0.0.12] - 2022-11-15

- Rename `GetBlockHash` to `GetBlockHashAtHeight`
- Reduce visibility scope of `IndexTransactions` to `indexTransactions`
- Remove `quorumCertificate` from the local context state
- Remove `LatestQC` and `SetLatestQC`
- Remove `Latest` prefix from several functions including related to setting context of the proposal block
- Added `ReleaseWriteContext` placeholder
- Replaced `ResetContext` with `Release`

## [0.0.0.11] - 2022-11-08

- Changed the following exported functions to lowercase non-exported functions
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

## [0.0.0.10] - 2022-11-01

- Ported over storing blocks and block components to the Persistence module from Consensus and Utility modules
- Encapsulated `TxIndexer` logic to the persistence context only

## [0.0.0.9] - 2022-10-19

- Fixed `ToPersistenceActors()` by filling all structure fields
- Deprecated `BaseActor` -> `Actor`
- Changed default actor type to `ActorType_Undefined`

## [0.0.0.8] - 2022-10-12

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`
- Made `PersistenceModule` struct unexported
- Updated tests and mocks
- Removed some cross-module dependencies
- Added `TxIndexer` sub-package (previously in Utility Module)
- Added `TxIndexer` to both `PersistenceModule` and `PersistenceContext`
- Implemented `TransactionExists` and `StoreTransaction`

## [0.0.0.7] - 2022-10-06

- Don't ignore the exit code of `m.Run()` in the unit tests
- Fixed several broken unit tests related to type casting

## [0.0.0.6] - 2022-09-30

- Removed no-op `DeleteActor` code
- Consolidated `CHANGELOG`s into one under `persistence/docs`
- Consolidated `README`s into one under `persistence/docs`
- Deprecated `persMod.ResetContext()` for -> `persRWContext.ResetContext()` for more appropriate encapsulation
- Added ticks to CHANGELOG.md
- Removed reference to Utility Mod's `BigIntToString()` and used internal `BigIntToString()`

## [0.0.0.5] - 2022-09-14

- Consolidated `PostgresContext` and `PostgresDb` into a single structure

## [0.0.0.4] - 2022-08-25

**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**

- Renamed schema -> types
- Added genesis, config, and unstaking proto files from shared
- Ensured proto structures implement shared interfaces
- Populate `PersistenceGenesisState` uses shared interfaces in order to accept `MockPersistenceGenesisState`
- ^ Same applies for `PersistenceConfig`
- Bumped cleanup TODOs to #149 due to scope size of #163

## [0.0.0.3] - 2022-08-16

**Main persistence module changes:**

- Split `ConnectAndInitializeDatabase` into `connectToDatabase` and `initializeDatabase`
  - This enables creating multiple contexts in parallel without re-initializing the DB connection
- Fix the SQL query used in `SelectActors`, `SelectAccounts` & `SelectPools`
  - Add a generalized unit test for all actors
- Remove `NewPersistenceModule` and an injected `Config` + `Create`
  - This improves isolation a a “injection-like” paradigm for unit testing
- Change `SetupPostgresDocker` to `SetupPostgresDockerPersistenceMod`
  - This enables more “functional” like testing by returning a persistence module and avoiding global testing variables
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

# Added

- Base persistence module implementation for the following actors: `Account`, `Pool`, `Validator`, `Fisherman`, `Servicer`, `Application`
- Generalization of common protocol actor behvaiours via the `ProtocolActor` and `BaseActor` interface and implementation
- A PostgreSQL based implementation of the persistence middleware including:
  - SQL query implementation for each actor
  - SQL schema definition for each actor
  - SQL execution for common actor behaviours
  - Golang interface implementation of the Persistence module
- Update to the Persistence module interface to enable historical height queries
- Library / infrastructure for persistence unit fuzz testing
- Tests triggered via `make test_persistence`

<!-- GITHUB_WIKI: changelog/persistence -->
