e2e tests
=========

> tl; dr - Run `make test_e2e` to run the e2e tests. 

## e2e with Docker Spike

This is a spike on using the Cucumber / Gherkin test suite to assert against `testcontainers-go`.  The goal is to expose test suite to a fresh network each time. Initially, this will be N=1 network configs, but future extension can happen here under the assumption. Of course, building the entire network is costly so scaling here could grow test times significantly.

### Realm 
The Realm maintains a list of nodes as a network. It exposes that network to the entire Cucumber test suite, then it cleans up that node network. Nodes are docker containers started from `github.com/testcontainers-go/tesetcontainers` that we run commands against.

### PocketClient
The Docker containers fulfill the PocketClient command interface. This interface accepts a broad string input for args and returns a `CommandResult`. This code is copied over from v0 `e2e/launcher`.

## Implementation Details
Using testcontainers-go at all is a serious choice here. The other detail we must choose is how to provision the Docker containers in a reliable way and then use them in our e2e suite. Should we use a custom docker-compose or try to rely on the compose_and_watch command and a side-car debug container?