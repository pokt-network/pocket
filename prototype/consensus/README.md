# Consensus Prototype Notes

## Usage

### [Optional] Run neo4j

```
$ mkdir -p neo4j/data
$ sudo chmod -R 777 neo4
$ make neo_d
```

### Run the nodes

Shell #1: `$ make compose_and_watch_no_neo`

### Run the client

Shell #2: `$ make client`

Keep calling one of the commands.

## High Level Plan

### High Level Tasks

[x] Reference existing implementation
[x] Start hacking/coding to get a feel for things
[ ] Split implementation into feature requests
[ ] Design a playbook for E2E tests (automatic or CLI driven)

### Ongoing in parallel to the above:

[ ] Update the pre-planning specification. Eventually turn it into a polished shareable design doc / whitepaper with the public.
[ ] Minimal unit tests where appropriate.
[ ] Document integration & E2E tests that need to be implemented.

## Components

[ ] Leader election logic
[ ] Blockchain Protocol
[ ] Consensus Protocol
[ ] Tooling (CLI and other)
[ ] Telemetry
[ ] Specification

### Leader Election Subcomponents

Components

[ ] VRF
[ ] CDF
[ ] Round Robin

### Blockchain Protocol

Components:

[ ] Genesis block
[ ] Types & interfaces
[ ] Evidence validation
[ ] Block validation

Testing:

- Okay use case for unit tests

### Consensus Protocol

Components:

- Types & interface
- Pacemaker
- Basic Hotstuff

Testing:

- Unit tests might be hard and cumbersome
- E2E Tests

### Tooling

Components:

[x] Code is already dockerized.
[x] Hot reload implemented
[o] CLI utility can be expanded

### Telemetry

Components:

[ ] Get centralized logging (loki) and metrics (grafana) working locally
[o] neo4j?
[ ] Network benchmarking out of the box

## Questions

### Specification

Q: When do we want to polish and release it?
A: original docs to be released in January and a final version after implementation is complete.

### Questions specific to consensus prototype

Q: How much P2P implementation do we want to have for development until the module is ready?
A: Logic has been decoupled and depdencies are being documented here: https://www.notion.so/Module-Interfaces-b0f4e2f8c3234d629079287747d81097

Q: How do we select shared libraries (e.g. crypto) for?
A: ???

Q: How to improve logging so we can distinguish between LEADER and REPLICA?
A: ???

### Questions with scope outside of just the consensus prototype

Q: Library for logging (to have levels)?
A: ???

Q: Opinion on use of `log.Fatal` or `log.Panic` vs just `log.Println` for errors?
A: ???

# Big questions to think about?

1. How do we we easily sift and filter through the logs (leader, replica, step, etc...)?
2. How can we visualize the message passing? Should be "interactive"?
3. How can we make a testing configuration language?
4. Can we visualize a dump of the logs?
5. How do we pause a round mid-way?

## References

## Pocket

- [Pocket 1.0 Consensus Module Pre-Planning Specification](https://docs.google.com/document/d/1bqsWmoztj_3JmKraEeZ1JMMBHkfW9k09FjVibHK2pd8)

## Github Repos:

- [hot-stuff/libhotstuff](https://github.com/hot-stuff/libhotstuff)
- [relab/hostuff](https://github.com/relab/hotstuff)
- [wjbbig/go-hotstuff](https://github.com/wjbbig/go-hotstuff)

## Academic Papers

- [HotStuff: BFT Consensus in the Lens of Blockchain](https://arxiv.org/pdf/1803.05069.pdf)
- [Twins: White-Glove Approach for BFT Testing](https://arxiv.org/pdf/2004.10617.pdf)
- [Fast-HotStuff: A Fast and Robust BFT Protocol for Blockchains](https://arxiv.org/pdf/2010.11454.pdf)
- [On the Performance of Pipelined HotStuff](https://arxiv.org/pdf/2107.04947.pdf)
