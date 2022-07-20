# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.1] - 2021-07-05

Pocket Persistence 1st Iteration (https://github.com/pokt-network/pocket/pull/73)

# Added

- Base persistence module implementation for the following actors: `Account`, `Pool`, `Validator`, `Fisherman`, `ServiceNode`, `Application`
- Generalization of common protocol actor behvaiours via the `ProtocolActor` and `BaseActor` interface and implementation
- A PostgreSQL based implementation of the persistence middleware including:
  - SQL query implementation for each actor
  - SQL schema definition for each actor
  - SQL execution for common actor behaviours
  - Golang interface implementation of the Persistence module
- Update to the Persistence module interface to enable historical height queries
- Library / infrastructure for persistence unit fuzz testing
- Tests triggered via `make test_persistence`
