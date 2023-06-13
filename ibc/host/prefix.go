package host

type (
	// CommitmentPrefix is the prefix bytes used in conjuction with a path string to create a
	// CommitmentPath. The prefix represents the store in which the path is stored
	CommitmentPrefix []byte
	// CommitmentPath is the path bytes used to store a commitment in a store
	CommitmentPath []byte
)

// ApplyPrefix applies the prefix to the provided path returning a CommitmentPath
func ApplyPrefix(prefix CommitmentPrefix, path string) CommitmentPath {
	bz := make([]byte, 0, len(prefix)+1+len([]byte(path)))
	bz = append(bz, prefix...)
	bz = append(bz, []byte("/")...)
	bz = append(bz, []byte(path)...)
	return CommitmentPath(bz)
}

// RemovePrefix removes the prefix from the provided CommitmentPath returning a path string
func RemovePrefix(prefix CommitmentPrefix, path CommitmentPath) string {
	return string(path[len(prefix)+1:])
}
