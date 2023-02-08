# Node Finite State Machine

The following diagram displays the various states and events that govern the functionality of the node.

```mermaid
stateDiagram-v2
    [*] --> stopped
    Consensus_syncMode --> Consensus_synced: Consensus_isCaughtUp
    Consensus_syncMode_serverMode --> Consensus_unsynched: disableServerMode
    Consensus_synced --> serverMode: enableServerMode
    Consensus_unsynched --> Consensus_syncMode: Consensus_isSyncing
    Consensus_unsynched --> Consensus_syncMode_serverMode: enableServerMode
    P2P_bootstrapped --> Consensus_synced: Consensus_isCaughtUp
    P2P_bootstrapped --> Consensus_unsynched: Consensus_isUnsynched
    P2P_bootstrapping --> P2P_bootstrapped: P2P_isBootstrapped
    serverMode --> Consensus_synced: disableServerMode
    stopped --> P2P_bootstrapping: start
```