# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.9] - 2023-02-06

- Documentation and supporting logic to enable `p1 debug` to be used from localhost

## [0.0.0.8] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.7] - 2023-02-04

- Changed log lines to utilize new logger module.

## [0.0.0.6] - 2023-02-02

- Fix broken link to `shared/crypto/README.md` in keybase documentation

## [0.0.0.5] - 2023-02-02

- Create `Keybase` interface to handle CRUD operations for `KeyPairs` with a `BadgerDB` backend
- Add logic to create, import, export, list, delete and update (passphrase) key pairs
- Add logic to sign and verify arbitrary messages
- Add unit tests for the keybase

## [0.0.0.4] - 2023-01-10

- The `client` (i.e. CLI) no longer instantiates a `P2P` module along with a bus of optional modules. Instead, it instantiates a `client-only` `P2P` module that is disconnected from consensus and persistence. Interactions with the persistence & consensus layer happen via RPC.
- Replaced previous implementation, reliant on `ValidatorMap`, with a temporary fetch from genesis. This will be replaced with a lightweight peer discovery mechanism in #416
- Simplified debug CLI initialization

## [0.0.0.3] - 2023-01-03

- Updated to use `coreTypes` instead of utility types for `Actor` and `ActorType`
- Updated README.md

## [0.0.0.2] - 2022-11-02

### Added

- Fixed message signing
- Reporting RPC StatusCode and body
- System commands working end-to-end
- Added Consensus State commands

## [0.0.0.1] - 2022-09-09

### Added

- Commands documentation generator

## [0.0.0.0] - 2022-09-07

### Added

- Basic implementation with Utility commands
  - Account
    - Send
  - Actor (Application, Node, Fisherman, Validator)
    - Stake (Custodial)
    - EditStake
    - Unstake
    - Unpause
  - Governance
    - ChangeParameter
  - Debug
    - Refactored previous CLI into a subcommand
- Functionally mocked a keybase in the form of a json file (default: pk.json) that will contain the privatekey
- CLI calling RPC via generated client
- Default configuration handling/overrides
