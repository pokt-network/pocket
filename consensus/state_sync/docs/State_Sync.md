## State Sync Lifecycle

Node starts syching with the rest of the network upon starting, as an external process via the `periodicMetaDataSynch()` function. Node keeps adding the metadata information it receives in the `syncMetadataBuffer` field of the stateSync struct via the `HandleStateSyncMetadataResponse()`.

Upon receiving a block, validator node checks the node's height, and if it's higher than it's current round, it checks if it is out of synch via `IsSynched()` function, which triggers aggregation of the metadata responses in the `syncMetadataBuffer` to the `aggregatedSyncMetadata`, compares with the current node height, and returns true if node is out of synch. In this case, node sends `StateMachineEvent_Consensus_IsUnsynched` event, which in turn, through FSM state transitions, triggers `StartSynching()` function. Or, if the node is in synch with its peers, then it rejects the block proposal. 

`StartSynching()` function requests block one by one using the minimum and maximum height info in the `aggregatedSyncMetadata` field, using the `broadCastStateSyncMessage()` function. The node, receives the requested blocks from the peers and applies the "correct" block to the persistence via `HandleGetBlockResponse()` function which checks quorum certificate and block validity.  
