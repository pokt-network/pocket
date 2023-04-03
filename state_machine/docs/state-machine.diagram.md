# Node Finite State Machine

The following diagram displays the various states and events that govern the functionality of the node.

```mermaid
stateDiagram-v2
    [*] --> Stopped
    Consensus_Pacemaker --> Consensus_Unsynced: Consensus_IsUnsynced
    Consensus_SyncMode --> Consensus_Synced: Consensus_IsSyncedNonValidator
    Consensus_SyncMode --> Consensus_Pacemaker: Consensus_IsSyncedValidator
    Consensus_Synced --> Consensus_Unsynced: Consensus_IsUnsynced
    Consensus_Unsynced --> Consensus_SyncMode: Consensus_IsSyncing
    P2P_Bootstrapped --> Consensus_Unsynced: Consensus_IsUnsynced
    P2P_Bootstrapping --> P2P_Bootstrapped: P2P_IsBootstrapped
    Stopped --> P2P_Bootstrapping: Start
```
