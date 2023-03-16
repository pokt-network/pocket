e2e tests
=========

> tl; dr - `make localnet_up` and then `make test_e2e` 

You can also click the `e2e-tests` button in the Tilt UI, which is handy during development.

## Design & Architecture

The test suite is located in `e2e/tests` and it contains a set of Cucumber Gherkin feature files and the associated Go tests to run them. This is a test file that lays out the BDD pattern using the client's help function as an example.

### Implementation 

`make test_e2e` sees any files named with the pattern `*.feature` in `e2e/tests` and runs them with [godog](https://github.com/cucumber/godog), the Go test runner for Cucumber tests. We indirectly depend on the LocalNet to be up and running for our test suite to use. We call commands to steer around a container by calling `kubectl exec` and targeting a pod in the cluster by name. We are able to pass commands to that client and then assert on the behavior of that program.

```
e2e/
└── tests
    ├── README.md
    ├── kube_client.go
    ├── root.feature
    └── steps_init_test.go

2 directories, 4 files
```

`kube_client.go` defines the KubeClient that grabs a Pod and drives it around for the tests.
`steps_init_test.go` registers the handler functions and runs the test suite.

### KubeClient and Kubernetes Clientsets

The KubeClient wraps the command and saves a reference to the last run command's result. The local environment's Kube client configuration is used to call commands. The Kubernetes Clientset can be changed in the test suite by swapping out the client that is defined at the package level. Multiple configurations can be maintained to test multiple environments and they can be swapped out arbitrarily.

### Separation of Concerns

The hardest architectural problem is correctly waiting for services so that we can call the commands at the correct moment, but doing so in the package without altering the LocalNet component. We should indirectly depend on the LocalNet but not modify it, since we consider it a separate concern. To achieve this, we will need to query the Kubernetes Clientset that we acquire from the local configuration and repeatedly poll that for changes until all services our tests rely on are running.

### Kubernetes Configurations

A goal of this project is to support swapping Kubernetes configurations out with the same test suite. This design accomplishes that by not requiring any direct dependency of the LocalNet and instead running commands through Kubernetes.