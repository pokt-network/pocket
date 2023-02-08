# Node Finite State Machine

The following diagram displays the various states and events that govern the functionality of the node.

```mermaid
stateDiagram-v2
    [*] --> stopped
    Consensus_syncMode --> Consensus_synced: Consensus_isCaughtUp
    Consensus_unsynched --> Consensus_syncMode: Consensus_isSyncing
    P2P_bootstrapped --> Consensus_synced: Consensus_isCaughtUp
    P2P_bootstrapped --> Consensus_unsynched: Consensus_isUnsynched
    P2P_bootstrapping --> P2P_bootstrapped: P2P_isBootstrapped
    stopped --> P2P_bootstrapping: start
```