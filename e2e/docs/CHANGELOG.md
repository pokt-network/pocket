# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.8] - 2023-05-31

- Adds the query feature file
- Renames the validator feature file
- Adds a Gherkin test for querying the block store through the Query Get CLI

## [0.0.0.7] - 2023-05-25

- Update E2E tests to use the `p1` binary name instead of the `client` binary name

## [0.0.0.6] - 2023-04-26

- Standardizes the kube pod name that E2E tests use

## [0.0.0.5] - 2023-04-24

- Attempts to fetch an in-cluster kubeconfig for E2E tests if none is found in `$HOME/.kube`

## [0.0.0.4] - 2023-04-19

- Changed validator DNS names to match new naming convention (again, helm chart was renamed)

## [0.0.0.3] - 2023-04-14

- Changed validator DNS names to match new naming convention
- Changed `RPC_HOST` default value to `pocket-validators` which randomly resolves to one of the validators

## [0.0.0.2] - 2023-04-10

Documentation updates

## [0.0.0.1] - 2023-04-10

Adds Stake, Unstake, & Send Tests [#653](https://github.com/pokt-network/pocket/pull/653)

- Introduced this `CHANGELOG.md`
- Added tests for Stake, Unstake, and Send CLI commands
- Added the `e2e/tests` directory

## [0.0.0.0] - 2023-03-30

Hello Changelog

<!-- GITHUB_WIKI: changelog/e2e -->
