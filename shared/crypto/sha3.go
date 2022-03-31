package crypto

import (
	"crypto"
	_ "golang.org/x/crypto/sha3"
)

var (
	hash        = crypto.SHA3_256
	SHA3HashLen = hash.Size()
)

func SHA3Hash(b []byte) []byte {
	hasher := hash.New()
	SHA3HashLen = crypto.SHA3_256.Size()
	return hasher.Sum(b)
}
