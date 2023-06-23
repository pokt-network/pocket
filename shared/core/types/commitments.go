package types

// These types are specific to the IBC module and how they store commitments in the
// IBC store. These types are used to implement the pre-defined paths from ICS-24.
// Ref: https://github.com/cosmos/ibc/blob/main/spec/core/ics-024-host-requirements/README.md
type (
	// CommitmentPrefix are the prefix bytes used in conjunction with a path string
	// to create a CommitmentPath.
	// The prefix represents the store in which the path is stored, for example:
	// => []byte("clients/{clientID}/consensusStates/{height}")
	// In the above example the prefix is []byte("clients") which represents the
	// client store, and the full byteslice is the CommitmentPath, the full key
	// under which the commitment is stored.
	CommitmentPrefix []byte
	// CommitmentPath is the path bytes used to store a commitment in a store
	CommitmentPath []byte
)
