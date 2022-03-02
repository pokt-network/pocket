package types

// The Pocket Network block height.
type BlockHeight uint64 // TODO: Move this into `types_consensus`.

// The number of times the node was interrupted at the current height; always 0 in the "happy path".
type Round uint8 // TODO: Move this into `types_consensus`.

type ConsensusNodeState struct {
	NodeId   NodeId `gogm:"name=NodeId"`
	Height   uint64 `gogm:"name=Height"` // TODO: Change to proper type
	Round    uint8  `gogm:"name=Round"`  // TODO: Change to proper type
	Step     uint8  `gogm:"name=Step"`   // TODO: Change to proper type
	IsLeader bool   `gogm:"name=IsLeader"`
	LeaderId NodeId `gogm:"name=Leader"`

	// gogm.BaseNode // Provides required node fields for neo4j DB
}
