package host

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// ApplyPrefix applies the prefix to the provided path returning a CommitmentPath
func ApplyPrefix(prefix coreTypes.CommitmentPrefix, path string) coreTypes.CommitmentPath {
	bz := make([]byte, 0, len(prefix)+1+len([]byte(path)))
	bz = append(bz, prefix...)
	bz = append(bz, []byte("/")...)
	bz = append(bz, []byte(path)...)
	return coreTypes.CommitmentPath(bz)
}

// RemovePrefix removes the prefix from the provided CommitmentPath returning a path string
func RemovePrefix(prefix coreTypes.CommitmentPrefix, path coreTypes.CommitmentPath) string {
	return string(path[len(prefix)+1:])
}
