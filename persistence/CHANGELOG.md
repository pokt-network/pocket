# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.3] - 2022-08-16

**Main persistence module changes:**

- Split `ConnectAndInitializeDatabase` into `connectToDatabase` and `initializeDatabase`
  - This enables creating multiple contexts in parallel without re-initializing the DB connection
- Fix the SQL query used in `SelectActors`, `SelectAccounts` & `SelectPools`
  - Add a generalized unit test for all actors
- Remove `NewPersistenceModule` and an injected `Config` + `Create`
  - This improves isolation a a “injection-like” paradigm for unit testing
- Change `SetupPostgresDocker` to `SetupPostgresDockerPersistenceMod`
  - This enables more “functional” like testing by returning a persistence module and avoiding global testing variables
  - Only return once a connection to the DB has been initialized reducing the likelihood of test race conditions
- Implemented `NewReadContext` with a proper read-only context

**Secondary persistence module changes**

- Improve return values in `Commit` and `Release` (return error, add logging, etc…)
- Add `pgx.Conn` pointer to `PostgresDB`
- `s/db/conn/g` and `s/conn/tx/g` in some (not all) places where appropriate
- Make some exported variables / functions unexported for readability & access purposes
- Add a few helpers for persistence related unit testing
- Added unit tests and TODOs for handling multiple read/write contexts

## [0.0.0.2] - 2022-08-03

Deprecate PrePersistence

- Fix for bytes parameters
- Accounts / pools default to 0
- Pre-added accounts to genesis file
- Separated out Persistence Read Context from Persistence Write Context
- Added various TODO's in order to code-complete a working persistence module
- Added genesis level functions to GetAllActors() and GetAllAccounts/Pools() for testing
- Added PopulateGenesisState function to persistence module
- Fixed the stake status iota issue
- Discovered and documented (with TODO) double setting parameters issue
- Attached to the Utility Module and using in `make compose_and_watch`

## [0.0.0.1] - 2022-07-05

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
