# Checklist to upgrade Go version

A short guide for carrying out Go version upgrades to Pocket V1

## Previous upgrades

A list of upgrades from the past, which can be used as a reference.

* 1.20 upgrade: [#910](https://github.com/pokt-network/pocket/pull/910)

## File Locations

- [ ]  go.mod
- [ ]  build.mk
- [ ]  Makefile
- [ ]  README.md
- [ ]  .golangci.yml
- [ ]  .github/workflows
    - [ ]  main.yml
    - [ ]  golangci-lint.yml
- [ ]  build/
    - [ ]  Dockerfile.client
    - [ ]  Dockerfile.debian.dev
    - [ ]  Dockerfile.debian.prod
    - [ ]  Dockerfile.localdev
- [ ]  docs/development
    - [ ]  README.md

## Testing

- [ ]  LocalNet builds and runs locally
- [ ]  LocalNet E2E tests pass
- [ ]  GitHub Actions CI tests pass
- [ ]  Remote network (such as DevNet) is functional and E2E tests pass
- [ ]  Update this document with current Pocket Go version