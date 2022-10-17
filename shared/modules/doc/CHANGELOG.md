# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.3] - 2022-10-16

Interface changes (w/o implementation) to support a StateHash computation.

### Persistence Module Changes

- Updates to the `PersistenceModule` interface

  - Introduce `ReleaseWriteContext`
  - Deprecate `ResetContext`

- Updates to the `PersistenceContext` interface

  - Change `Commit()` to `Commit(proposerAddr, quorumCert)`
  - Change `AppHash()` to `UpdateAppHash()`
  - Deprecate `StoreBlock`, `InsertBlock`, `Reset`

- Updates to the `UtilityContext` interface
  - Change `ReleaseContext()` to `Release()`
  - Changed `CommitPersistenceContext()` to `Commit(quorumCert)`
  - Deprecate `GetPersistenceContext`

## [0.0.0.2] - 2022-09-17

- Minimized shared module with [#163](https://github.com/pokt-network/pocket/issues/163)
- Deprecated shared/types, moved remaining interfaces to shared/modules
- Most GenesisTypes moved to persistence

## [0.0.0.1] - 2022-08-21

- Minimized shared module with [#163](https://github.com/pokt-network/pocket/issues/163)
- Deprecated shared/types, moved remaining interfaces to shared/modules
- Most GenesisTypes moved to persistence

## [0.0.0.0] - 2022-08-08

- Deprecated old placeholder genesis_state and genesis_config
- Added utility_genesis_state to genesis_state
- Added consensus_genesis_state to genesis_state
- Added genesis_time to consensus_genesis_state
- Added chainID to consensus_genesis_state
- Added max_block_bytes to consensus_genesis_state
- Added accounts and pools to utility_genesis_state
- Added validators to utility_genesis_state
- Added applications to utility_genesis_state
- Added service_nodes to utility_genesis_state
- Added fishermen to utility_genesis_state
- Deprecated shared/config/
- Added new shared config proto3 structure
- Added base_config to config
- Added utility_config to config
- Added consensus_config to config
- Added persistence_config to config
- Added p2p_config to config
- Added telemetry_config to config
- Opened followup issue #163
- Added config and genesis generator to build package
- Deprecated old build files
- Use new config and genesis files for make compose_and_watch
- Use new config and genesis files for make client_start && make client_connect
