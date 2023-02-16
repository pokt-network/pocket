# Node Finite State Machine

The following diagram displays the various states and events that govern the functionality of the node.

```mermaid
stateDiagram-v2
    [*] --> Stopped
    Consensus_SyncMode --> Consensus_Synced: Consensus_IsCaughtUp
    Consensus_Unsynched --> Consensus_SyncMode: Consensus_IsSyncing
    P2P_Bootstrapped --> Consensus_Synced: Consensus_IsCaughtUp
    P2P_Bootstrapped --> Consensus_Unsynched: Consensus_IsUnsynched
    P2P_Bootstrapping --> P2P_Bootstrapped: P2P_IsBootstrapped
    Stopped --> P2P_Bootstrapping: Start
```