# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0] - 2022-09-30

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Abstracted config and genesis handling
- Mockable runtime
- Refactored all modules to use RuntimeMgr
- Updated RuntimeMgr to manage clock as well
- Modules now accept `interfaces` instead of paths.
- Unmarshalling is done in a new `runtime` package (runtime because what we do in there affects the runtime of the application)
- We are now able to accept configuration via environment variables (thanks to @okdas for inspiration and [sp13 for Viper]("github.com/spf13/viper"))
