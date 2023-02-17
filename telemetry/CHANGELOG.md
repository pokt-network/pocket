# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.8] - 2023-02-17

- Modules embed `base_modules.IntegratableModule` and `base_modules.InterruptableModule` for DRYness

## [0.0.0.7] - 2023-02-07

- Added GITHUB_WIKI tags where it was missing

## [0.0.0.6] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.5] - 2023-02-06

- Changed readme to remove $ sign from code blocks

## [0.0.0.4] - 2023-02-04

- Changed log lines to utilize new logger module.

## [0.0.0.3] - 2023-01-19

- Rewrite `interface{}` to `any`

## [0.0.0.2] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering module with the `bus` via `RegisterModule` method

## [0.0.0.1] - 2023-01-03

- Refactored configs into `configs` package

## [0.0.0.0] - 2023-01-03

- Introduced this `CHANGELOG.md`

<!-- GITHUB_WIKI: changelog/telemetry -->
