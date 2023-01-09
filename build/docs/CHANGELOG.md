# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.2] - 2023-01-09

- Reorder private keys so addresses (retrieved by transforming private keys) to reflect the numbering in LocalNet appropriately. The address for val1, based on config1, will have the lexicographically first address. This makes debugging easier.

## [0.0.0.1] - 2023-01-03

- Removed `BaseConfig` from `configs`
- Centralized `PersistenceGenesisState` and `ConsensusGenesisState` into `GenesisState`

## [0.0.0.0] - 2022-12-22

- Introduced this `CHANGELOG.md`
