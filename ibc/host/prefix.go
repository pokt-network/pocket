package host

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// ApplyPrefix applies the prefix to the provided path returning a CommitmentPath
func ApplyPrefix(prefix *coreTypes.CommitmentPrefix, path string) *coreTypes.CommitmentPath {
	bz := make([]byte, len(prefix.Prefix)+len([]byte(path)))
	bz = append(bz, prefix.Prefix...)
	bz = append(bz, []byte(path)...)
	return &coreTypes.CommitmentPath{Path: bz}
}

// RemovePrefix removes the prefix from the provided CommitmentPath returning a path byte slice
func RemovePrefix(prefix *coreTypes.CommitmentPrefix, path *coreTypes.CommitmentPath) []byte {
	return path.Path[len(prefix.Prefix):]
}
