# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.2] - 2023-01-04

- Removed `BaseConfig` from `configs`
- Centralized `PersistenceGenesisState` and `ConsensusGenesisState` into `GenesisState`

## [0.0.0.1] - 2022-12-29

- Updated configs with the missing value `max_mempool_count`
- Added `is_client_only` to `config1.json` so that Viper knows it can be overridden. Done in the Makefile in `make client_connect`. Setting this can be avoided if we merge the changes in https://github.com/pokt-network/pocket/compare/main...issue/cli-viper-environment-vars-fix

## [0.0.0.0] - 2022-12-22

- Introduced this `CHANGELOG.md`
