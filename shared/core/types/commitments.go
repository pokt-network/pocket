package types

type (
	// CommitmentPrefix is the prefix bytes used in conjunction with a path string to create a
	// CommitmentPath. The prefix represents the store in which the path is stored
	CommitmentPrefix []byte
	// CommitmentPath is the path bytes used to store a commitment in a store
	CommitmentPath []byte
)
