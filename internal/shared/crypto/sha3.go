package crypto

import (
	"crypto"
	"encoding/hex"

	_ "golang.org/x/crypto/sha3"
)

var (
	SHA3HashLen = crypto.SHA3_256.Size()
)

func SHA3Hash(bz []byte) []byte {
	hasher := crypto.SHA3_256.New()
	hasher.Write(bz)
	return hasher.Sum(nil)
}

// GetHashStringFromBytes returns the hexadecimal encoding in string format of the SHA3Hash of the bytes passed in as argument
//
// Typically used to compute a TransactionHash
func GetHashStringFromBytes(bytes []byte) string {
	return hex.EncodeToString(SHA3Hash(bytes))
}
