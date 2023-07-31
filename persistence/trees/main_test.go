//go:build test

package trees

import (
	"crypto/sha256"
	"hash"
)

type TreeStore = treeStore

var SMTTreeHasher hash.Hash = sha256.New()
