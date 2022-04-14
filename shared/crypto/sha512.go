package crypto

import (
	"crypto"
	"crypto/sha512"
)

var (
	SHA512HashLen = crypto.SHA512.Size()
)

func SHA512Hash(bz []byte) []byte {
	hasher := sha512.New()
	hasher.Write(bz)
	return hasher.Sum(nil)
}
