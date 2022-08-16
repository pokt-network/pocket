package raintree

import (
	hex "encoding/hex"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// CLEANUP: Consolidate this with other similar functions in the codebase.
func GetHashStringFromBytes(b []byte) string {
	hash := cryptoPocket.SHA3Hash(b)
	return hex.EncodeToString(hash)
}
