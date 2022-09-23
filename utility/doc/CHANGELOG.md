# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.4] - 2022-09-23
- Created UtilityConfig
- Added `Max_Mempool_Transaction_Bytes` and `Max_Mempool_Transactions` to the utility 
  config to allow dynamic configuration of the mempool
- Matched configuration unmarshalling pattern of other modules
- Added V0 mempool default configurations
- Regenerated build files with new mempool config

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
