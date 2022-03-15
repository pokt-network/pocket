package crypto

import (
	"crypto"
	_ "golang.org/x/crypto/sha3"
)

var (
	SHA3HashLen = crypto.SHA3_256.Size()
)

func SHA3Hash(b []byte) []byte {
	hasher := crypto.SHA3_256.New()
	hasher.Write(b)
	return hasher.Sum(nil)
}
