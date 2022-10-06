# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
