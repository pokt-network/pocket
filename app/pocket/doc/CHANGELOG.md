# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.3] - 2023-02-06

- Introduced `CONFIG_PATH` and `GENESIS_PATH` environment variables for debug CLI commands

## [0.0.0.2] - 2023-02-03

- Changed log lines to utilize new logger module.

## [0.0.0.1] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore

## [0.0.0.0] - 2022-11-02

### Added

- Added the simplest form of feature flagging for the RPC server functionality
- Calling the RPC server entrypoint in a goroutine if enabled
