package types

// The Pocket Network block height.
type BlockHeight uint64 // TODO: Move this into `consensus_types`.

// The number of times the node was interrupted at the current height; always 0 in the "happy path".
type Round uint8 // TODO: Move this into `consensus_types`.
