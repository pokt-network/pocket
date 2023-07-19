# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.36] - 2023-06-19

- Add a new trustless relay sub-command to servicer command

## [0.0.0.35] - 2023-06-14

- Update documentation and tests in the keybase to use the Secretbox key encryption method

## [0.0.0.34] - 2023-06-13

- Exported `rootCmd` to support cross-package usage (i.e. subcommands w/ own pkgs)
- Refactored common CLI code
  - Moved & refactored `PersistentPreRunE()` helper
  - Moved & exported `busCLICtxKey`
  - Exported `GetValueFromCLIContext()` and `SetValueInCLIContext()`
  - Moved `setupAndStartP2PModule`
  - Moved `setupCurrentHeightProvider`
  - Moved `setupPeerstoreProvider`
- Refactored CLI flags to own package for cross-package use
- Replaced RPC_HOST with POCKET_REMOTE_CLI_URL where appropriate
- Added support for overriding of `remote-cli-url` flag with env var
- Improve error handling in client CLI

## [0.0.0.33] - 2023-06-13

- Renamed `NewRPCPeerstoreProvider()` and `NewPersistencePeerstoreProvider()` to `Create()` (per package)

## [0.0.0.32] - 2023-05-25

- Add the `nonInteractive` flag in a couple spots where it was missing

## [0.0.0.31] - 2023-05-25

- Change user facing cli name to `p1`

## [0.0.0.30] - 2023-05-04

- Add query subcommands to interact with a node's RPC server

## [0.0.0.29] - 2023-04-28

- Adds debug subcommands matching list of interactive prompt commands

## [0.0.0.28] - 2023-04-17

- Refactor debug CLI post P2P module re-consolidation

## [0.0.0.27] - 2023-04-07

- Add Query Command
- Add AllChainParams subcommand to query governance parameters

## [0.0.0.26] - 2023-03-30

- Make `PromptPrintNodeState` the first prompt in debug mode
- Minor cleanup to documentation related to CLI modes
- Fixed one logging statement

## [0.0.0.25] - 2023-03-28

- Introduces hashicorp vault keybase to allow for the use of a vault server to store keypairs

## [0.0.0.24] - 2023-03-28

- Automatic import reorder

## [0.0.0.23] - 2023-03-21

- Refactor debug CLI to use new P2P interfaces

## [0.0.0.22] - 2023-03-17

- Added a limit on concurrent key imports for debug client to avoid OOM.

## [0.0.0.21] - 2023-03-14

- Simplifies the debug CLI tooling by embedding private-keys.yaml manifest
  into the CLI binary when the debug build tag is present.

## [0.0.0.20] - 2023-03-03

- Support libp2p module in debug CLI

## [0.0.0.19] - 2023-02-28

- Renamed the package names for some basic helpers

## [0.0.0.18] - 2023-02-28

- Implement SLIP-0010 HD child key derivation with the keybase
- Add CLI endpoints to derive child keys by index

## [0.0.0.17] - 2023-02-23

- Add CLI endpoints to interact with the keybase

## [0.0.0.16] - 2023-02-21

- Rename ServiceNode Actor Type Name to Servicer

## [0.0.0.15] - 2023-02-17

- Added `non_interactive` flag to allow for non-interactive `Stake` and `Unstake` transactions (dogfooding in `cluster-manager`)
- Updated CLI to use to source the address book and the current height from the RPC server leveraging the `rpcAddressBookProvider` and `rpcCurrentHeightProvider` respectively and the `bus` for dependency injection

## [0.0.0.14] - 2023-02-15

- Introduced logical switch to handle parsing of the debug private keys from a local file OR from Kubernetes secret (PR #517)
- Bugfix for `Stake` command. Address erroneously sent instead of the PublicKey. (PR #518)

## [0.0.0.13] - 2023-02-14

- Fixed `docgen` to work from the root of the repository
- Updated all the CLI docs

## [0.0.0.12] - 2023-02-14

- Integrate keybase with CLI
- Add debug module to keybase to automatically populate keybase with 999 validators

## [0.0.0.11] - 2023-02-09

- Added debugging prompts to drive state sync requests
- `SendMetadataRequest` to send metadata request by all nodes to all nodes
- `SendBlockRequest` to send get block request by all nodes to all nodes

## [0.0.0.10] - 2023-02-07

- Added GH_WIKI tags where it was missing

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

- Fixed message signing
- Reporting RPC StatusCode and body
- System commands working end-to-end
- Added Consensus State commands

## [0.0.0.1] - 2022-09-09

- Commands documentation generator

## [0.0.0.0] - 2022-09-07

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

<!-- GITHUB_WIKI: changelog/client -->
