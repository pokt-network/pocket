# E2E Testing Framework

<!-- TOC -->

- [E2E Testing Framework](#e2e-testing-framework)
  - [Implementation](#implementation)
  - [Kubernetes Config](#kubernetes-config)
  - [Validator Tests](#validator-tests)

<!-- /TOC -->


> tl; dr - `make localnet_up` and then `make test_e2e`

![Tilt button for running E2E tests](docs/tilt-button.png)

You can also click the `e2e-tests` button in the Tilt UI, which is handy during development.

## Implementation

The test suite is located in `e2e/tests` and it contains a set of Cucumber feature files and the associated Go tests to run them. `make test_e2e` sees any files named with the pattern `*.feature` in `e2e/tests` and runs them with [godog](https://github.com/cucumber/godog), the Go test runner for Cucumber tests. The LocalNet must be up and running for the E2E test suite to run.

## Kubernetes Config

This test suite assumes that you have LocalNet setup and have a `~/.kube/config` file that connects to it. That config is loaded up and used to retrieve a Clientset which allows us deep access in to the LocalNet.

## Validator Tests

The Validator calls RPC commands on the container by calling `kubectl exec` and targeting the pod in the cluster by name. It records the results of the command including stdout and stderr, allowing for assertions about the results of the command.

## Build Tags

Because the E2E tests depend on a Kubernetes environment to be available, the E2E tests package gets a build tag so the E2E tests are ignored unless the test command is run with -tags=e2e. Issue [#581](https://github.com/pokt-network/pocket/issues/581) covers running the E2E tests in the Github Actions pipeline.
