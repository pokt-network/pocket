# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.6] - 2022-10-12
- Stores transactions alongside blocks during `commit`
- Added current block `[]TxResult` to the module

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
