# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.14] - 2023-02-08

- Updated all `config*.json` files with new `server_mode_enabled` field (for state sync)

## [0.0.0.13] - 2023-02-08

- Fix bug related to installing Tilt in the Docker containers

## [0.0.0.12] - 2023-02-07

- Code formatting by VSCode

## [0.0.0.11] - 2023-02-07

- Added GITHUB_WIKI tags where it was missing

## [0.0.0.10] - 2023-02-06

- Added `genesis_localhost.json`, a copy of `genesis.json` to be used by the localhost instead of a debug docker container

## [0.0.0.9] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.8] - 2023-02-06

- Added LocalNet on Kubernetes with tilt.dev

## [0.0.0.7] - 2023-02-04

- Added `--decoration="none"` flag to `reflex`

## [0.0.0.6] - 2023-01-23

- Added pprof feature flag guideline in docker-compose.yml

## [0.0.0.5] - 2023-01-20

- Update the docker-compose and relevant files to automatically load `pgadmin` server configs by binding the appropriate configs

## [0.0.0.4] - 2023-01-14

- Added `max_conns_count`, `min_conns_count`, `max_conn_lifetime`, `max_conn_idle_time` and `health_check_period` to config files

## [0.0.0.3] - 2023-01-11

- Reorder private keys so addresses (retrieved by transforming private keys) to reflect the numbering in LocalNet appropriately. The address for val1, based on config1, will have the lexicographically first address. This makes debugging easier.

## [0.0.0.2] - 2023-01-10

- Removed `BaseConfig` from `configs`
- Centralized `PersistenceGenesisState` and `ConsensusGenesisState` into `GenesisState`
- Removed `is_client_only` since it's set programmatically in the CLI

## [0.0.0.1] - 2022-12-29

- Updated all `config*.json` files with the missing `max_mempool_count` value
- Added `is_client_only` to `config1.json` so Viper knows it can be overridden. The config override is done in the Makefile's `client_connect` target. Setting this can be avoided if we merge the changes in https://github.com/pokt-network/pocket/compare/main...issue/cli-viper-environment-vars-fix

## [0.0.0.0] - 2022-12-22

- Introduced this `CHANGELOG.md`

<!-- GITHUB_WIKI: changelog/build -->
