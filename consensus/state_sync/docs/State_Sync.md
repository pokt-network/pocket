## State Sync Lifecycle

Node starts syching with the rest of the network upon starting, as an external process via the `periodicMetaDataSynch()` function. Node keeps adding the metadata responses to aggregates the metadata information it receives in the `syncMetadataBuffer` field of the stateSync struct.

Upon receiving a block, validator node checks the node's height, and if it's higher than it's current round, it checks if it is out of synch via `IsSynched()` function, which triggers aggregation of the metadata responses in the `syncMetadataBuffer`, compares with the current node height, and returns true if node is out of synch. In this case, node sends `StateMachineEvent_Consensus_IsUnsynched` event. Which in turn, through FSM state transitions, triggers `StartSynching()` function. 