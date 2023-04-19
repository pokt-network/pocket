# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.53] - 2023-04-18

- Added a new `Session` protobuf
- Added a `GetActor` function to the Persistence module interface
- Added a `GetSession` function to the Utility module interface

## [0.0.0.52] - 2023-04-17

- Removed _temporary_ `shared/p2p` package; consolidated into `p2p`

## [0.0.0.51] - 2023-04-13

- Consolidate all the `TxResult` protobufs and interfaces into a common protobuf located in `shared/core/types`

## [0.0.0.50] - 2023-04-10

- Added `modules.ModuleFactoryWithOptions` interface
- Added factory interfaces:
  - `modules.FactoryWithRequired`
  - `modules.FactoryWithOptions`
  - `modules.FactoryWithRequiredAndOptions`
- Embedded `ModuleFactoryWithOptions` in `Module` interface
- Switched mock generation to use reflect mode for effected interfaces (embedders)

## [0.0.0.49] - 2023-04-07

- Removed `GetParameter()` from `PersistenceReadContext`
- Add `gov_utils.go` to create a map of all metadata related to governance parameters

## [0.0.0.48] - 2023-04-07

- Renamed `CreateAndApplyProposalBlock` to `CreateProposalBlock`
- Added `GetStateHash` to `UtilityUnitOfWork`

## [0.0.0.47] - 2023-04-06

- Updated to reflect pools address changes
- Added tests to catch, in a future-proof way, changes to our pools
- Updated interfaces to use `[]byte` instead of `string` for `pool` addresses for consistency with `accounts` and because otherwise fuzzy tests would fail

## [0.0.0.46] - 2023-04-03

- Add `ConsensusStateSync` interface. It defines exported state sync functions in consensus module
- Update `ConsensusDebugModule` with getter and setter function for state sync testing
- Update FSM events `Consensus_IsSyncedValidator`, `Consensus_IsSyncedNonValidator` and state `Consensus_Pacemaker`

## [0.0.0.45] - 2023-03-30

- Add a deadline to the primary event handling to get visibility into concurrency issues

## [0.0.0.44] - 2023-03-28

- Add UnmarshalJSON to KeyPair to unmarshal public key correctly

## [0.0.0.43] - 2023-03-26

- Updated `PROTOCOL_STATE_HASH.md` to reference the `UtilityUnitOfWork`
- Refactored interfaces to use `UtilityUnitOfWork`
- Added interfaces for `UtilityUnitOfWork` and `UtilityUnitOfWorkFactory`
- Added interfaces `LeaderUtilityUnitOfWork` and `ReplicaUtilityUnitOfWork`
- Updated `UtilityModule` to use `UtilityUnitOfWork`
- Refactored utility module implementation to use `UtilityUnitOfWork` and moved into separate sub-package

## [0.0.0.42] - 2023-03-21

- Add `TransitionEventToMap()` helper function for logging

## [0.0.0.41] - 2023-03-21

- Added _temporary_ `shared/p2p` package to hold P2P interfaces common to both legacy and libp2p modules
- Added `Peerstore` interface
- Added `Peer` and `PeerList` and interfaces
- Moved `typesP2P.AddrBookMap` to `sharedP2P.PeerAddrMap` and refactor to implement the new `Peerstore` interface
- Refactored `getAddrBookDelta` to be a member of `PeerList`
- Factored `SortedPeerManager` out of `raintree.peersManager` and add `PeerManager` interface
- Refactored getAddrBookDelta to be a member of PeerList

## [0.0.0.40] - 2023-03-20

- Adds enum DebugMessageType for distinguishing message routing behavior

## [0.0.0.39] - 2023-03-09

- Fix diagrams in SLIP documentation to be in the correct order

## [0.0.0.38] - 2023-03-03

- Support libp2p module in node

## [0.0.0.37] - 2023-03-01

- add pokt --> libp2p crypto helpers

## [0.0.0.36] - 2023-02-28

- Move `StakeStatus` into `actor.proto`
- Rename `generic_param` into `service_url`
- Remove unused params from `BlockHeader`
- Document `transaction.proto` and move it from the `utility` module to `shared`
- Moved `signature.go` from the `utility` module to `shared`
- Added documentation to important functions in the `Persistence` and `Utility` shared modules
- Added documentation on how/why `UnstakingActor` should be removed

## [0.0.0.35] - 2023-02-28

- Implement SLIP-0010 specification for HD child key derivation
- Cover both Test Vectors 1 and 2 from the specification

## [0.0.0.34] - 2023-02-24

- Remove SetLeaderId() method from ConsensusDebugModule interface

## [0.0.0.33] - 2023-02-24

- Update logger value references with pointers

## [0.0.0.32] - 2023-02-23

- Add file utility functions for checking, reading and writing to files

## [0.0.0.31] - 2023-02-22

- Export consensus module's ConsensusDebugModule interface.

## [0.0.0.30] - 2023-02-21

- Rename ServiceNode Actor Type Name to Servicer

## [0.0.0.29] - 2023-02-20

- Fan-ing out `StateMachineTransitionEventType` event to the `P2P` module to handle bootstrapping logic
- Refactored single key generation from seed (used in tests) into `GetPrivKeySeed`

## [0.0.0.28] - 2023-02-17

- Added `UnmarshalText` to `Ed25519PrivateKey`
- Fan-ing out `ConsensusNewHeightEventType` events

## [0.0.0.27] - 2023-02-17

- Added events `ConsensusNewHeightEvent` and `StateMachineTransitionEvent`
- Introduced `BaseInterruptableModule` and `IntegratableModule` to reduce repetition and boilerpate code (DRYness)
- Added `ModulesRegistry` and `StateMachineModule` accessors and interfaces
- Introduced generic `ModuleOption` pattern to fine tune modules behaviour
- Added `StateMachine` to the `node` initialization

## [0.0.0.26] - 2023-02-16

- Added `FetchValidatorPrivateKeys` function since it is going to be used by the `debug-client` and also by the upcoming `cluster-manager` [#490](https://github.com/pokt-network/pocket/issues/490)

## [0.0.0.25] - 2023-02-14

- Remove shared `ActorTypes` array and use the enum directly
- Reduce the code footprint of the `codec` package & add some TODOs
- Added `UnstakingActor` proto to remove deduplication across modules; adding TECHDEBT to remove altogether one day
- Added clarifying comments to the utility module interface

## [0.0.0.24] - 2023-02-09

- Add `ConsensusStateSync` interface that is implemented by the consensus module

## [0.0.0.23] - 2023-02-07

- Added GITHUB_WIKI tags where it was missing

## [0.0.0.22] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.21] - 2023-02-04

- Changed log lines to utilize new logger module.
- Added example to Readme how to initiate a logger using new logger module.

## [0.0.0.20] - 2023-02-03

- Introduced `GenericFIFOList` to handle generic FIFO mempool lists (can contain duplicates)
- Introduced `GenericFIFOSet` to handle generic FIFO mempool sets (items are unique)
- Updated `Utility` module interface to expose mempool access via `GetMempool()`

## [0.0.0.19] - 2023-02-02

- Add `KeyPair` interface
- Add logic to create new keypairs, encrypt/armour them and decrypt/unarmour them

## [0.0.0.18] - 2023-01-31

- Match naming conventions in `Param` protobuf file

## [0.0.0.17] - 2023-01-27

- Add `Param` and `Flag` protobufs for use in updating merkle tree

## [0.0.0.16] - 2023-01-24

- Add `ConsensusPacemaker` interface that is implemented by the consensus module

## [0.0.0.15] - 2023-01-20

- Remove `address []byte` argument from `InsertPool` function in `PostgresRWContext`

## [0.0.0.14] - 2023-01-19

- Rewrite `interface{}` to `any`

## [0.0.0.13] - 2023-01-18

- Create `block.proto` which consolidates the definition of a `Block` protobuf under `shared/core/types`

## [0.0.0.12] - 2023-01-11

- Deprecated `GetBlocksPerSession()` and `GetServicersPerSessionAt()` in favour of the more general parameter getter function `GetParameter()`

## [0.0.0.11] - 2023-01-11

- Make the events channel hold pointers rather than copies of the message

## [0.0.0.10] - 2023-01-10

- Updated modules constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering modules with the `bus` via `RegisterModule` method

## [0.0.0.9] - 2023-01-04

- Removed `ValidatorMap() ValidatorMap` from `ConsensusModule` interface
- Added `GetIsClientOnly()` to `P2PConfig`

## [0.0.0.8] - 2023-01-03

- Added `PoolNames.FriendlyName` method
- Renamed enums as per code-review
- Updated `InitParams` logic to use genesisState instead of hardcoded values

## [0.0.0.7] - 2022-12-21

- Updated to use the new centralized config and genesis handling
- Created `Actor` struct under `coreTypes`
- Created `Account` struct under `coreTypes`
- Created `PoolNames` enum under `coreTypes`
- Updated module to use the new `coreTypes`
- Simplified `*Module` interfaces
- Updated tests and mocks

## [0.0.0.6] - 2022-12-14

- Added `GetMaxMempoolCount`

## [0.0.0.5] - 2022-12-06

- Change the `bus` to be a pointer receiver rather than a value receiver in all the functions it implements

## [0.0.0.4] - 2022-11-30

Debug:

- `ResetToGenesis` - Added the ability to reset the state to genesis
- `ClearState` - Added the ability to clear the state completely (height 0 without genesis data)

Configs:

- Updated the test generator to produce deterministic keys
- Added `trees_store_dir` to persistence configs
- Updated `LocalNet` configs to have an empty `tx_indexer_path` and `trees_store_dir`

## [0.0.0.3] - 2022-11-14

### [#353](https://github.com/pokt-network/pocket/pull/353) Remove topic from messaging

- Removed topic from messaging
- Added `PocketEnvelope` as a general purpose wrapper for messages/events
- Added utility methods to `Pack` and `Unpack` messages
- Replaced the switch cases, interfaces accordingly

## [0.0.0.2] - 2022-10-12

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`, available via `GetRuntimeMgr()`
- Segregate interfaces (eg: `GenesisDependentModule`, `P2PAddressableModule`, etc)
- Updated tests and mocks

## [0.0.0.1] - 2022-09-30

- Used proper `TODO/INVESTIGATE/DISCUSS` convention across package
- Moved TxIndexer Package to Utility to properly encapsulate
- Add unit test for `SharedCodec()`
- Added `TestProtoStructure` for testing
- Flaky tests troubleshooting - https://github.com/pokt-network/pocket/issues/192
- More context here as well: https://github.com/pokt-network/pocket/pull/198

### [#198](https://github.com/pokt-network/pocket/pull/198) Flaky tests

- Time mocking abilities via https://github.com/benbjohnson/clock and simple utility wrappers
- Race conditions and concurrency fixes via sync.Mutex

## [0.0.0.0] - 2022-08-25

### [#163](https://github.com/pokt-network/pocket/issues/163) Minimization

- Moved all shared structures out of the shared module
- Moved structure responsibility of config and genesis to the respective modules
- Shared interfaces and general 'base' configuration located here
- Moved make client code to 'debug' to clarify that the event distribution is for the temporary local net
- Left multiple `TODO` for remaining code in test_artifacts to think on removal of shared testing code

<!-- GITHUB_WIKI: changelog/shared -->
