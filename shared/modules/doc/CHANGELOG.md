# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.7] - 2023-01-11

- Added comments to the functions exposed by `P2PModule`

## [0.0.0.6] - 2022-12-10

Persistence Module:

- Add `proposerAddr` input to the `Commit` function
- Remove `SetProposalBlock`, `GetBlockTxs` and `GetProposerAddr`
- Rename `ComputeAppHash` to `ComputeStateHash`

Utility Module:

- Introduce the `SetProposalBlock` function

## [0.0.0.5] - 2022-12-07

- Changed the scope of `TransactionExists` from the `PostgresContext` to the `PersistenceModule`

## [0.0.0.4] - 2022-11-30

- Removed `GetPrevHash` and just using `GetBlockHash` instead
- Removed `blockProtoBz` from `SetProposalBlock` interface
- Removed `GetLatestBlockTxs` and `SetLatestTxResults` in exchange for `IndexTransaction`
- Removed `SetTxResults`
- Renamed `UpdateAppHash` to `ComputeStateHash`
- Removed some getters related to the proposal block (`GetBlockTxs`, `GetBlockHash`, etc…)

## [0.0.0.3] - 2022-11-15

PersistenceModule

- Added `ReleaseWriteContext`
- Consolidated `ResetContext`, `Reset` with `Release`
- Modified `Commit` to accept a `quorumCert`
- Removed `Latest` prefix from getters related to the proposal block parameters

UtilityModule

- Changed `CommitPersistenceContext()` to `Commit(quorumCert)`
- Changed `ReleaseContext` to `Release`

## [0.0.0.2] - 2022-10-12

- Modified interface for Utility Module `ApplyBlock` and `GetProposalTransactions` to return `TxResults`
- Modified interface for Persistence Module `StoreTransaction` to store the `TxResult`
- Added shared interface `TxResult` under types.go

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
- Added servicers to utility_genesis_state
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

<!-- GITHUB_WIKI: changelog/shared_modules -->
