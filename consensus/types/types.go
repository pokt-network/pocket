package types

// The Pocket Network block height.
type BlockHeight uint64 // TODO: Move this into `types_consensus`.

// The number of times the node was interrupted at the current height; always 0 in the "happy path".
type Round uint8 // TODO: Move this into `types_consensus`.

type NodeId uint64

type ConsensusNodeState struct {
	NodeId   NodeId
	Height   uint64
	Round    uint8
	Step     uint8
	IsLeader bool
	LeaderId NodeId
}
