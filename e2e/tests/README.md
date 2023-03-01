e2e tests 
=========

> tl; dr - Run `make test_e2e` to run the e2e tests. 

## Network with Docker Spike

This is a spike on using the Cucumber / Gherkin test suite with docker testcontainers-go.  The goal is to expose test suite to a fresh network each time.  Initially, this will be N=1 network configs, but future extension can happen here under the assumption.  Of course, building the entire network is costly so scaling here could grow test times significantly.

### Network 
Network maintains a list of nodes as a network. It exposes that network to the entire Cucumber test suite, then it cleans up that node network.
Nodes are docker containers started from `github.com/testcontainers-go/tesetcontainers` that we run commands against.