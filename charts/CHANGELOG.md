# Changelog

All notable changes to helm charts will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.2] - 2023-04-17

- Removed `runtime/configs.Config#UseLibp2p` field
- Set validator `POCKET_P2P_HOSTNAME` env var to the pod IP
- Set validator `p2p.hostname` config value to empty string so that the env var applies

## [0.0.0.1] - 2023-04-14

- Introduced `pocket-validator` helm chart.

<!-- GITHUB_WIKI: changelog/charts -->