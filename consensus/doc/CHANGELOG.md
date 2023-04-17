# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.48] - 2023-04-17

- Debug logging improvements

## [0.0.0.47] - 2023-04-17

-  Log warnings in `handleStateSyncMessage()` 

## [0.0.0.46] - 2023-04-13

- Utilise the `TxResult` protobuf from `shared/core/types`

## [0.0.0.45] - 2023-04-12

- Added new helper functions for testing hotstuff consensus stages: `waitForNewRound()`, `waitForPrepareProposal()`, `waitForPrepareVoteswaitForPreCommit()`, `waitForCommit()`, and `waitForDecide()` , and a wrapper function `waitForNextBlock()`.
- Removed state machine mock in testing, and introduced actual state machine that starts with `StartAllTestPocketNodes()` function and added `generatePlaceholderBlock()` function
- Added, new helper functions for testing state sync stages: `waitForNodeToRequestMissingBlock()`, `waitForNodeToReceiveMissingBlock()`, `waitForNodeToCatchUp()` and a wrapper function `waitForNodeToSync()` currently as placeholders.

## [0.0.0.44] - 2023-04-12

- Updated `handleStateSyncMessage()` to log warnings if server mode is not enabled

## [0.0.0.43] - 2023-04-07

- Renamed `CreateAndApplyProposalBlock` to `CreateProposalBlock`
- Updated to reflect the new `ApplyBlock` signature from `utilityUnitOfWork`
- Updated to use `utilityUnitOfWork.GetStateHash()`
- Removed duplicate import

## [0.0.0.42] - 2023-04-03

- Add `fsm_handler.go` to handle FSM transition events in consensus module
- Update State Machine mock in `utils_test.go`
- Update state_sync module with additional function definitions

## [0.0.0.41] - 2023-03-30

- Improve & simplify `utilityUnitOfWork` management
- Logging - improve the wording and context of various logging statements
- Logging - remove a lot of unused logging function helpers
- Genesis - Fix ordering of operations when resetting to genesis
- Added `msgToLoggingFields(hotstuffMsg)` helper
- Added missing `persistenceContext.Release()` calls
- Avoid redundant handling of hotstuff messages by replicas
- Consolidated `prev` & `last` terminology
- Consolidated `stateHash` & `prevHash` terminology
- Removed unused `logPrefix` variables
- Moved implementation of `modules.ConsensusDebugModule` into its own file
- Moved implementation of `modules.ConsensusPacemaker` into its own file
- Moved implementation of `modules.ConsensusStateSync` into its own file

## [0.0.0.40] - 2023-03-29

- Bugfix for (unable to send txs in localnet) #631

## [0.0.0.39] - 2023-03-26

- Refactored `utilityContext` into `utilityUnitOfWork`
- Added `utilityUnitOfWorkFactory` to create `utilityUnitOfWork` instances depending on the fact that the current node is `Leader` or `Replica`
- Renamed `prepareAndApplyBlock` to `prepareBlock`
- Centralized `applyBlock` logic
- Updated tests

## [0.0.0.38] - 2023-03-25

- Add quorum certificate to the block before committing to persistence
- Add error `ErrNoQcInReceivedBlock` and message `DisregardBlock`

## [0.0.0.37] - 2023-03-21

- Add helper function `logHelper()` for logging

## [0.0.0.36] - 2023-02-28

- Creating a persistence read context when needing to accessing stateless (i.e. block hash) data
- Renamed package names and parameters to reflect changes in the rest of the codebase
- Removed the unused `validator.proto`

## [0.0.0.35] - 2023-02-28

- Fixed bug in `sendGetMetadataStateSyncMessage`

## [0.0.0.34] - 2023-02-24

- Fixed `TestPacemakerCatchupSameStepDifferentRounds` test

## [0.0.0.33] - 2023-02-24

- Update logger value references with pointers

## [0.0.0.32] - 2023-02-22

- Move functions exposed by ConsensusDebugModule to debugging.go

## [0.0.0.31] - 2023-02-21

- Rename ServiceNode Actor Type Name to Servicer

## [0.0.0.30] - 2023-02-17

- Updated log messages in the state sync submodule with consistent style and add height information
- Added state sync message types to the types package

## [0.0.0.29] - 2023-02-17

- Modules embed `base_modules.IntegratableModule` and `base_modules.InterruptableModule` for DRYness
- Updated modules `Create` to accept generic options
- `resetToGenesis` clears the utility mempool as well
- Publishing `ConsensusNewHeightEvent` on new height

## [0.0.0.28] - 2023-02-14

- Add a few `nolint` comments to fix the code on main

## [0.0.0.27] - 2023-02-09

- Add `state_sync` submodule, with `state_sync` struct
- Implement state sync server to advertise blocks and metadata
- Create new `state_sync_handler.go` source file that handles `StateSyncMessage`s sent to the `Consensus` module
- Add two new tests in `state_sync_test.go`:`TestStateSyncServerGetMetaDataReq` and `TestStateSyncServerGetBlock`
- Update `TestHotstuff4Nodes1BlockHappyPath` test to also retrieve the committed block

## [0.0.0.26] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.25] - 2023-02-04

- Changed log lines to utilize new logger module.

## [0.0.0.24] - 2023-02-03

- Introduced `hotstuffFIFOMempool` that extends the logic provided by the genericized FIFO mempool in `shared`.
- Added tests for `hotstuffFIFOMempool` to ensure that it behaves as expected.

## [0.0.0.23] - 2023-01-30

- Fix `TestHotstuff4Nodes1BlockHappyPath` misplacement of actual and expected values in `require.Equal`

## [0.0.0.22] - 2023-01-25

- Add a note on consensus test related workaround related to #462

## [0.0.0.21] - 2023-01-24

- Decouple consensus module and pacemaker module
- Add `pacemaker` submodule
- Update pacemaker struct to remove consensus module field, and related functions
- Create new `pacemaker_consensus.go` source file that consists ConsensusPacemaker function implementations

## [0.0.0.20] - 2023-01-19

- Rewrite `interface{}` to `any`

## [0.0.0.19] - 2023-01-18

- Remove `Block` proto definition to consolidate under `shared/core/types`

## [0.0.0.18] - 2023-01-11

### Consensus - Core

- Force consensus to use a "star-like" broadcast instead of "RainTree" broadcast
- Improve logging throughout through the use of emojis and rewording certain statements
- Slightly improve the block verification flow (renaming, minor fixes, etc…) to stabilize LocalNet

### Consensus - Tests

- Rename the `consensus_tests` package to `e2e_tests`
- Internalize configuration related to `fail_on_extra_msgs` from the `Makefile` to the `consensus` module
- Forced all tests to fail if we receive extra unexpected messages and modify tests appropriately
- After #198, we made tests deterministic but there was a hidden bug that modified how the test utility functions because the clock would not move while we were waiting for messages. This prevented logs from streaming, tests from failing, and other issues. Tend to all related changes.

### Consensus - Pacemaker

- Rename `ValidateMessage` to `ShouldHandleMessage` and return a boolean
- Pass a `reason` to `InterruptRound`
- Improve readability of some parts of the code

## [0.0.0.17] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering module with the `bus` via `RegisterModule` method
- Updated tests and mocks accordingly

## [0.0.0.16] - 2023-01-09

- Added protobuf message definitions for requests related to sharing state sync metadata and blocks
- Defined the interface for `StateSyncServerModule`, `StateSyncModule` (moving the old interface to `StateSyncModuleLEGACY` as a reference only)
- Overhaul (updates, improvements, clarifications & additions) of the State Sync README
- Removed `ValidatorMap() ValidatorMap`

## [0.0.0.15] - 2023-01-03

- ValidatorMap uses `Actor` references now

## [0.0.0.14] - 2022-12-21

- Updated do use the new centralized config and genesis
- `Actor` is now a shared `struct` instead of an `interface`
- Removed converters between the interfaces and the consensus structs for Validators

## [0.0.0.13] - 2022-12-14

- Consolidated number of validators in tests in a single constant: `numValidators`
- Fixed typo in `make test_consensus_concurrent_tests` so that we can run the correct test matrix
- Using `GetBus()` instead of `bus` wherever possible
- `LeaderElectionModule`'s `electNextLeaderDeterministicRoundRobin` now uses `Persistence` to access the list of validators instead of the static `ValidatorMap`.

## [0.0.0.12] - 2022-12-12

- Unexport `ConsensusModule` fields
- Create `ConsensusDebugModule` interface with setter functions to be used only for debugging puroposes
- Update test in `TestPacemakerCatchupSameStepDifferentRounds` in `pacemaker_test.go` to use setter functions

## [0.0.0.11] - 2022-12-06

- Removed unused `consensus.UtilityMessage`

## [0.0.0.10] - 2022-11-30

- Propagate `highPrepareQC` if available to the block being created
- Remove `blockProtoBytes` from propagation in `SetProposalBlock`
- Guarantee that write context is released when refreshing the utility context
- Use `GetBlockHash(height)` instead of `GetPrevAppHash` to be more explicit
- Use the real `quorumCert` when preparing a new block

## [0.0.0.9] - 2022-11-30

- Added state sync interfaces and diagrams

## [0.0.0.8] - 2022-11-15

- Propagate the `quorumCertificate` on `Block` commit to the `Utility` module
- Slightly improved error handling of the `utilityContext` lifecycle management

## [0.0.0.7] - 2022-11-01

- Removed `apphash` and `txResults` from `consensusModule` structure
- Modified lifecycle to `set` the proposal block within a `PersistenceContext`
- Allow block and parts to be committed with the persistence context

## [0.0.0.6] - 2022-10-12

- Stores transactions alongside blocks during `commit`
- Added current block `[]TxResult` to the module

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`
- Made `ConsensusModule` struct unexported
- Updated tests and mocks
- Removed some cross-module dependencies

## [0.0.0.5] - 2022-10-06

- Don't ignore the exit code of `m.Run()` in the unit tests

## [0.0.0.4] - 2022-09-28

- `consensusModule` stores block directly to prevent shared structure in the `utilityModule`

## [0.0.0.3] - 2022-09-26

Consensus logic

- Pass in a list of messages to `findHighQC` instead of a hotstuff step
- Made `CreateProposeMessage` and `CreateVotemessage` accept explicit values, identifying some bugs along the way
- Made sure to call `applyBlock` when using `highQC` from previous round
- Moved business logic for `prepareAndApplyBlock` into `hotstuff_leader.go`
- Removed `MaxBlockBytes` and storing the consensus genesis type locally as is

Consensus cleanup

- Using appropriate getters for protocol types in the hotstuff lifecycle
- Replaced `proto.Marshal` with `codec.GetCodec().Marshal`
- Reorganized and cleaned up the code in `consensus/block.go`
- Consolidated & removed a few `TODO`s throughout the consensus module
- Added TECHDEBT and TODOs that will be require for a real block lifecycle
- Fixed typo in `hotstuff_types.proto`
- Moved the hotstuff handler interface to `consensus/hotstuff_handler.go`

Consensus testing

- Improved mock module initialization in `consensus/e2e_tests/utils_test.go`

General

- Added a diagram for `AppHash` related `ContextInitialization`
- Added `Makefile` keywords for `TODO`

## [0.0.0.2] - 2022-08-25

**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**

- Ensured proto structures implement shared interfaces
- `ConsensusConfig` uses shared interfaces in order to accept `MockConsensusConfig` in test_artifacts
- `ConsensusGenesisState` uses shared interfaces in order to accept `MockConsensusGenesisState` in test_artifacts
- Implemented shared validator interface for `validator_map` functionality

## [0.0.0.1] - 2021-03-31

### Added new libraries: HotPocket 1st iteration

- Initial implementation of Basic Hotstuff
- Initial implementation Hotstuff Pacemaker
- Deterministic round robin leader election
- Skeletons, passthroughs and temporary variables for utility integration
- Initial implementation of the testing framework
- Tests with `make test_pacemaker` and `make test_hostuff`

## [0.0.0.0] - 2021-03-31

### Added new libraries: VRF & Cryptographic Sortition Libraries

- Tests with `make test_vrf` and `make test_sortition`
- Benchmarking via `make benchmark_sortition`
- VRF Wrapper library in `consensus/leader_election/vrf/` of github.com/ProtonMail/go-ecvrf/ecvrf
- Implementation of Algorand's Leader Election sortition algorithm in `consensus/leader_election/sortition/`

<!-- GITHUB_WIKI: changelog/consensus -->
