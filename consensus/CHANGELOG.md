# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.2] - 2022-08-25
**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**
- Ensured proto structures implement shared interfaces
- `ConsensusConfig` uses shared interfaces in order to accept `MockConsensusConfig` in test_artifacts
- `ConsensusGenesisState` uses shared interfaces in order to accept `MockConsensusGenesisState` in test_artifacts
- Implemented shared validator interface for `validator_map` functionality

## [0.0.0.1] - 2021-03-31

HotPocket 1st Iteration (https://github.com/pokt-network/pocket/pull/48)

# Added

- Initial implementation of Basic Hotstuff
- Initial implementation Hotstuff Pacemaker
- Deterministic round robin leader election
- Skeletons, passthroughs and temporary variables for utility integration
- Initial implementation of the testing framework
- Tests with `make test_pacemaker` and `make test_hostuff`

## [0.0.0.0] - 2021-03-31

VRF & Cryptographic Sortition Libraries (https://github.com/pokt-network/pocket/pull/37/files)

### Added

- Tests with `make test_vrf` and `make test_sortition`
- Benchmarking via `make benchmark_sortition`
- VRF Wrapper library in `consensus/leader_election/vrf/` of github.com/ProtonMail/go-ecvrf/ecvrf
- Implementation of Algorand's Leader Election sortition algorithm in `consensus/leader_election/sortition/`
