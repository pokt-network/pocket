# Consensus Module <!-- omit in toc -->

This document is meant to be a supplement to the living specification of [1.0 Pocket's Consensus Module Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/consensus) primarily focused on the implementation, and additional details related to the design of the codebase.

## Table of Contents <!-- omit in toc -->

- [Interface](#interface)
- [Consensus Module Lifecycle](#consensus-module-lifecycle)
  - [Leader Election](#leader-election)
  - [Block Generation Process](#consensus-phases)
  - [Block Validation Process](#block-validation-process)
  - [State Sync Process](#state-sync-process)
- [Implementation](#implementation)
  - [Code Organization](#code-organization)
- [Testing](#testing)
  - [Running Unit Tests](#running-unit-tests)

## Interface

This module aims to implement the interface specified in `pocket/shared/modules/consensus_module.go` using the specification above.


## Consensus Module Overview

This repository features an implementation of the HotStuff consensus algorithm. Consensus process is facilitated through a series of rounds. Staked validator nodes participate in the consensus process, where one node is elected as the leader, and the others act as replicas.

### Leader Election

Leader election is handled by a dedicated submodule. In our current configuration, we utilize a deterministic round-robin leader election mechanism as the primary leader election method.

### Consensus Phases

The HotStuff consensus algorithm has three phases: `Prepare`, `Pre-Commit`, and `Commit`. In each phase, the leader creates a proposal and broadcasts it to all replica nodes.

Upon receiving the proposal, each replica node performs block validation check. If the proposal is valid, the replica node responds to the leader with its signature, which acts as its vote.

Once the leader collects votes from more than two-thirds of the replicas, it moves on to the next consensus phase. This two-thirds rule is critical for satisfying the Byzantine Fault Tolerance (BFT) requirement, ensuring the network's resilience against faulty or malicious nodes.


### Block Generation Process
```mermaid
sequenceDiagram
    participant Leader
    participant Replicas
    Note over Leader,Replicas: Leader Election
    Leader->>Replicas: Propose(block)
    Note over Replicas: Validate proposed block
    Replicas-->>Leader: Prepare(block)
    Note over Leader: Receives Prepare messages from a quorum of Replicas
    Leader->>Replicas: Pre-Commit(block, Prepare messages)
    Note over Replicas: Validate Pre-Commit message
    Replicas-->>Leader: Commit(block)
    Note over Leader: Receives Commit messages from a quorum of Replicas
    Leader->>Replicas: Notify(block, Commit messages)
    Note over Replicas: Add block to local blockchain copy
    Note over Leader,Replicas: New Leader Election
```

### Block Validation Process
```mermaid
graph TD
    A[Receive Block Proposal from Leader]
    B[Check Block Structure]
    C[Check Block Hash]
    D[Check Previous Block Reference]
    E[Check Transactions]
    F[Check Block Creator's Signature]
    G[Check Timestamp]
    H[Block is Valid - Proceed with Prepare message]
    I[Block is Invalid - Reject Block]
    A-->B
    B-->C
    C-->D
    D-->E
    E-->F
    F-->G
    G-->H
    B-.->I
    C-.->I
    D-.->I
    E-.->I
    F-.->I
    G-.->I
```


### State Syncronisation

State synchronization is an essential process in our consensus module to ensure that all participating nodes maintain an up-to-date and consistent view of the network state. It is particularly important in a dynamic and decentralized network environment where nodes can join or leave, or might be intermittently offline. For in-depth understanding of the state sync and current status check out our [State Sync Protocol Design Specification](https://github.com/pokt-network/pocket/blob/main/consensus/doc/PROTOCOL_STATE_SYNC.md).


```mermaid
graph TD
    A(Start testing) --> Z(Add new validators)
    Z --> B[Trigger Next View]
    B --> C{BFT threshold satisfied?}
    C -->|Yes| D(New block, height increases)
    C -->|No| E(No new block, height is same)
    E --> B
    D --> F{Are there new validators staked?}
    F -->|Yes| G(Wait for validators' metadata responses)
    F -->|No| J{Are syncing nodes catched up?}
    J --> |Yes| Z
    J -->|No| B
    G --> B

    subgraph Notes
       note1>NOTE: BFT requires > 2/3 validators<br>in same round & height, voting for proposal.]
       note2>NOTE: Syncing validators request blocks from the network.]
    end

    C --> note1
    J --> note2
```


### Consensus Lifecycle

```mermaid
flowchart TD
  A[Start New Round] --> |Elect Leader| L[Leader Election Module]
  L --> D1[Leader]
  L --> D2[Replica]
  D1 --> E1[Create Proposals]
  D2 --> E2[Validate Proposals]
  E1 --> F1[Aggregate Votes]
  E2 --> F2[Vote on Proposals]
  F1 --> G1[Quorum and Commit Block]
  F2 --> G2[Commit Block]
  G1 --> J1[End Round]
  G2 --> J1
  J1 --> A
```



## Implementation

### Code Organization

```bash
consensus
├── doc
│   ├── CHANGELOG.md                        
│   ├── PROTOCOL_STATE_SYNC.md              # State sync protocol definition
├── e2e_tests
│   ├── hotstuff_test.go                    # Hotstuff consensus tests
│   ├── pacemaker_test.go                   # Pacemaker module tests
│   ├── state_sync_test.go                  # State sync tests
│   ├── utils_test.go                       # test utils
├── leader_election                         
│   ├── sortition                           
│       └── sortition_test.go               # Sortition tests
│       └── sortition.go                    # Cryptographic sortition implementation
│   ├── vrf                                 
│       └── errors.go                       
│       └── vrf_test.go                     # VRF tests
│       └── vrf.go                          # VRF implementation
│   ├── module.go                           # Leader election module implementation
├── pacemaker                                  
│   ├── debug.go                            
│   ├── module.go                           # Pacemaker module implementation
├── state_sync                                 
│   ├── helpers.go                          
│   ├── interfaces.go                       
│   ├── module.go                           # State sync module implementation
│   ├── server.go                           # State sync server functions
├── telemetry   
│   ├── metrics.go                          
├── types
│   ├── proto                               # Proto3 messages for generated types
│   ├── actor_mapper_test.go
│   ├── actor_mapper.go           
│   ├── messages.go                         # Consensus message definitions 
│   ├── types.go                            # Consensus type definitions
├── block.go                                 
├── debugging.go                            # Debug function implementation
├── events.go                                
├── fsm_handler.go                          # FSM events handler implementation
├── helpers.go                              
├── hotstuff_handler.go                     
├── hotstuff_leader.go                      # Hotstuff message handlers for Leader
├── hotstuff_mempool_test.go                # Mempool tests
├── hotstuff_mempool.go                     # Hotstuff transaction mempool implementation
├── hotstuff_replica.go                     # Hotstuff message handlers for Replica
├── messages.go                             # Hotstuff message helpers
├── module_consensus_debugging.go            
├── module_consensus_pacemaker.go           # Pacemaker module helpers
├── module_consensus_state_sync.go          # State sync module helpers
├── module.go                               # The implementation of the Consensus Interface
├── README.md                               # Self link to this README
├── state_sync_handler.go                   # State sync message handler
```

## Testing

### Running Unit Tests

```bash
make test_consensus
```