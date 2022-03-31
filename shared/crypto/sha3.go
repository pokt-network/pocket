package crypto

import (
	"crypto"
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
