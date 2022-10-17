# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.3] - 2022-10-16

- Updates to the `PersistenceModule` interface
  - Added `ReleaseWriteContext`
  - Removed `ResetContext`
- Updates to the `PersistenceContext` interface
  - Removed `Reset`
  - Changed `AppHash` `UpdateAppHash`
  - Changed `Commit()` to `Commit(proposerAddr, quorumCert)`
- Updates to the `UtilityContext` interface
  - Change `ReleaseContext` to `Release`
  - Removed `GetPersistenceContext`
  - Changed `CommitPersistenceContext()` to `Commit(quorumCert)`

## [0.0.2] - 2022-10-12

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`, available via `GetRuntimeMgr()`
- Segregate interfaces (eg: `GenesisDependentModule`, `P2PAddressableModule`, etc)
- Updated tests and mocks

## [0.0.1] - 2022-09-30

- Used proper `TODO/INVESTIGATE/DISCUSS` convention across package
- Moved TxIndexer Package to Utility to properly encapsulate
- Add unit test for `SharedCodec()`
- Added `TestProtoStructure` for testing
- Flaky tests troubleshooting - https://github.com/pokt-network/pocket/issues/192
- More context here as well: https://github.com/pokt-network/pocket/pull/198

### [#198](https://github.com/pokt-network/pocket/pull/198) Flaky tests

- Time mocking abilities via https://github.com/benbjohnson/clock and simple utility wrappers
- Race conditions and concurrency fixes via sync.Mutex

## [0.0.0] - 2022-08-25

### [#163](https://github.com/pokt-network/pocket/issues/163) Minimization

- Moved all shared structures out of the shared module
- Moved structure responsibility of config and genesis to the respective modules
- Shared interfaces and general 'base' configuration located here
- Moved make client code to 'debug' to clarify that the event distribution is for the temporary local net
- Left multiple `TODO` for remaining code in test_artifacts to think on removal of shared testing code
