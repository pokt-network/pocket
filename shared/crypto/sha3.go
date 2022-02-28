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
	return hasher.Sum(b)
}
