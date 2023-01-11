# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.17] - 2023-01-11

- Deprecated `GetBlocksPerSession()` in favour of the more general parameter getter function `GetParameter()`
- Update unit test for `GetBlocksPerSession()` to use the `GetParameter()` function

## [0.0.0.16] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering module with the `bus` via `RegisterModule` method

## [0.0.0.15] - 2023-01-03

- Renamed enum names as per code-review
- Using defaults from `test_artifacts` for tests
- Updated tests to reflect the above changes

## [0.0.0.14] - 2022-12-21

- Updated to use the new centralized config and genesis handling
- Updated to use the new `Actor` struct under `coreTypes`
- Updated tests and mocks

## [0.0.0.13] - 2022-12-10

- Introduce `SetProposalBlock` and local vars to keep proposal state
- Maintaining proposal block state (proposer, hash, transactions) in local context

## [0.0.0.12] - 2022-12-06

- Introduce a general purpose `HandleMessage` method at the utility level
- Move the scope of `CheckTransaction` from the context to the module level
- Add an `IsEmpty` function to the `Mempool`
- Rename `DeleteTransaction` to `RemoveTransaction` in the mempool
- Rename `LatestHeight` to `Height` in the `utilityContext`
- Add comments inside `CheckTransaction` so its functionality is clearer
- Add comments and cleanup the code in `mempool.go`

## [0.0.0.11] - 2022-11-30

- Minor lifecycle changes needed to supported the implementation of `ComputeAppHash` as a replacement for `GetAppHash` in #285

## [0.0.0.10] - 2022-11-15

- Propagating the `quorumCertificate` appropriately on block commit
- Removed `Latest` from getters related to retrieving the context of the proposed block

## [0.0.0.9] - 2022-11-01

- Remove `TxResult` from the utility module interface (added in TxIndexer integration of transaction indexer (issue-#168) #302)
- Combined creation and application of block in proposer lifecycle

## [0.0.0.8] - 2022-10-17

- Added Relay Protocol interfaces and diagrams

## [0.0.0.7] - 2022-10-14

- Added session interfaces and diagrams
- Moved `TxIndexer` package to persistence module
- Added new proto structure `DefaultTxResult`
- Integrated the `TxIndexer` into the lifecycle
  - Captured `TxResult` from each played transaction
  - Moved the storage of transactions to the Consensus Module
  - Returned the `TxResults` in the `ApplyBlock()` and `GetProposalTransactions()`
  - `AnteHandleMessage()` now returns `signer`
  - `ApplyTransaction()` returns `TxResult`

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`
- Made `UtilityModule` struct unexported
- Updated tests and mocks
- Removed some cross-module dependencies

## [0.0.0.6] - 2022-10-06

- Don't ignore the exit code of `m.Run()` in the unit tests
- Fixed several broken unit tests related to type casting
- Removed some unit tests (e.g. `TestUtilityContext_UnstakesPausedBefore`) that were legacy and replaced by more general ones (e.g. `TestUtilityContext_UnstakePausedBefore`)
- Avoid exporting unnecessary test helpers

## [0.0.0.5] - 2022-09-29

- Remove unused `StoreBlock` function from the utility module interface

## [0.0.0.4] - 2022-09-23

- Created `UtilityConfig`
- Added `max_mempool_transaction_bytes` and `max_mempool_transactions` to the utility
  config to allow dynamic configuration of the mempool
- Matched configuration unmarshalling pattern of other modules
- Added V0 mempool default configurations
- Regenerated build files with new mempool config
- Consolidated `UtilActorType` in `utility` and `utility/types` to `typesUtil.ActorType`
- Deprecated all code in `actor.go` and replaced with test helpers
- Converted stake status to proto.enum (int32)
- Added DISCUSS items around shared code and `StakeStatus`
- Removed no-op `DeleteActor` code
- Improved unit test for `UnstakeActorsThatAreReady()`
- Removed all usages of `fmt.Sprintf()` from the testing package
- Replaced all usages of `requre.True/require.False` with `require.Equal` unless checking a boolean
- Added helper function for getting height and store for a readable and consistent `typesUtil.Error` value
- Added testing.M argument to `newTestingPersistenceModule`
- Moved in-function _literal_ arguments for `newTestingPersistenceModule` to private constants
- Added the address parameter to `ErrInsufficientFunds` function for easier debugging
- Added unit test for `LegacyVote.ValidateBasic()`
- Added `ErrUnknownActorType` to all switch statements on `actorType`
- Removed `import` of `consTypes` (consensus module)

## [0.0.0.3] - 2022-09-15

### Code cleanup

- Consolidated `TransactionHash` to call a single implementation in `shared/crypto/sha3`
- Extracted function calls from places where we were using the same logic

## [0.0.0.2] - 2022-08-25

**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**

- Ensured proto structures implement shared interfaces
- `UtilityConfig` uses shared interfaces in order to accept `MockUtilityConfig` in test_artifacts
- Moved all utilty tests from shared to tests package
- Left `TODO` for tests package still importing persistence for `NewTestPersistenceModule`
  - This is one of the last places where cross-module importing exists

## [0.0.1] - 2022-07-20

### Code cleanup

- Removed transaction fees from the transaction structure as fees will be enforced at the state level
- Removed actor specific messages (besides DoubleSign) and added actorType field to the struct
- Removed pause messages and functionality as it is out of scope for the current POS iteration
- Removed session and test-scoring as it's out of scope for the current POS iteration
- Consolidated unit test functionality for actors
- Modified pre-persistence to match persistence for Update(actor), 'amountToAdd' is now just 'amount'

## [Unreleased]

## [0.0.0] - 2022-03-15

### Added

- Added minimal 'proof of stake' implementation with few Pocket Specific terminologies and actors
  - Structures
    - Accounts
    - Validators
    - Fishermen
    - Applications
    - Service Nodes
    - Pools
  - Messages
    - Stake
    - Unstake
    - EditStake
    - Pause
    - Unpause
    - Send
- Added initial governance params
