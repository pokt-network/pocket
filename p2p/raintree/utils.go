package raintree

import (
	hex "encoding/hex"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

func GetHashStringFromBytes(b []byte) string {
	hash := cryptoPocket.SHA3Hash(b)
	return hex.EncodeToString(hash)
}
