# Usage

Shell 1

```
$ make protogen_local
$ make compose_and_watch
```

Shell 2

```
$ make client_start
$ make client_connect
```

Start starting new views.

# Notes (Protocol Hour 02/18/2022)

- The state validation is missing (_purposefully_) however the validators still produce blocks (_take into consideration there is no byzantine behavior, state is valid_)
  - Utility is there, however rigorous state validation is not enabled.
- We are not discarding the transactions from the mempool and all the shenanigans and the details of code-completeness, but we produce BLOCK 1 !!! 🥳
- We have hot-reloading on from the get go on 5 validators, and this is unbelievably productive!
- Code walkthrough
  - `cmd/pocket/main.go` is the entrypoint
    - this file loads a configuration from a file. (check `shared/config.go`)
    - the actual json configuration lives in `build/config/*.json` (including `genesis.json`)
    - the configuration has a path to genesis file, p2p configuration, node_id, and keys (private_key) and all the rest filler configurations for the module to come (persistence, utility...)
  - we got a `module.go` that exists in every existing module, where the actual module implements the Module interface, that will allow it to be usable from external parties, particularly by the node entrypoint (`v1/main.go`)
    - actually, the other modules extend the `Module` interface
    - the god module is `Node` struct, and it lives in `shared/node.go` and implements `Module` as well.
  - the node itself implements the Module interface, and in its `Create` method, it calls all the other modules’ `Create` methods
  - for the time being, we are using a state mock, instead of a true real genesis state. This mock state lives in `persistence/pre_persistence/test_state.go`
  - the way we instantiate the node is as follows:
    - create modules
    - create the bus (application specific bus)
    - inject it to other modules
    - and voila!
  - another important piece of the shared namespace is the `handler.go`
    - this piece is what allows the node to coordinate in between different modules through the bus (`GetBus` lives in there and allows the node to retrieve other modules functionality through the bus)
  - other capabilities of shared are:
    - crypto
    - types (got proto in)
  - another important one is shared/module
    - it hosts the singular modules interfaces definitions
      - i.e: utility defines the UtilityContext behavior in utility_module.go + the definition of the UtilityModule interface which extends the shared Module interface
      - i.e: persistence defines the PersistenceContext in persistence_module.go + the definition of the PersistenceModule interface which extends the shared Module interface
      - similarly, all other modules go like this
  - specific modules:
    - utility:
      - has its own types
      - has its own type of context
      - implements the Module interface in module.go (_of course_)
  - the currently used p2p module (pre_p2p) that acts a mock is a simplistic mock to allow consensus to move forward
    - it basically uses send to write on an established tcp connection and closes it immediately once done
    - the broadcast is a linear iteration over validators + a send
    - there is a handle network message that basically reads off of a connection and retrieves the received message and passes it to the proper module
  - consensus:
    - implements the Module interface and follows the same pattern for its submodules like the pacemaker, dkg, statesync and all the others such that:
      - consensus.Create calls on the module.Create of each peacemaker, dkg and statesync
      - consensus defines the overaching struct it needs and implements the module.go immediately in the module.go
      - although it’s a single module, files are separated by scope even if they implement the same interface, such as block.go for instance that implements a few new (private) methods on the ConsensusModule, but it is being done in a separate file.
