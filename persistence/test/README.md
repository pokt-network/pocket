## How to test

From the project root

- Command: `go test -p 1 persistence/test/...`

### Pre-Requisites
- Must run Docker Desktop (or similar Docker client) https://www.docker.com/products/docker-desktop/

### Notes
- There are many TODO's in the testing environment including thread safety. 
It's possible that running the tests in parallel may cause tests to break so it is recommended to use `-p 1` flag
