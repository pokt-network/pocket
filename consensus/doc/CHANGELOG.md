# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.14] - 2022-12-29

- Removed `ValidatorMap() ValidatorMap`

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

- Improved mock module initialization in `consensus/consensus_tests/utils_test.go`

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
