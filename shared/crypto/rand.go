package crypto

import (
	"crypto/ed25519"
	crand "crypto/rand"
	"encoding/binary"
	"math/big"
	"math/rand"
	"time"
)

const maxNonce = ^uint64(0)

// Generate cryptographically secure random nonce
func GetNonce() uint64 {
	max := new(big.Int)
	max.SetUint64(maxNonce)
	bigNonce, err := crand.Int(crand.Reader, max)
	if err != nil {
		// If failed to get cryptographically secure nonce use a pseudo-random nonce
		rand.Seed(time.Now().UTC().UnixNano()) //nolint:staticcheck // G404 - Weak random source is okay in unit tests
		return rand.Uint64()                   //nolint:gosec // G404 - Weak source of random here is fallback
	}

	// 0 is an invalid value
	if bigNonce.Uint64() == 0 {
		return GetNonce()
	}
	return bigNonce.Uint64()
}

// GetPrivKeySeed returns a private key from a seed
func GetPrivKeySeed(seed int) PrivateKey {
	seedBytes := make([]byte, ed25519.PrivateKeySize)
	binary.LittleEndian.PutUint32(seedBytes, uint32(seed))
	pk, err := NewPrivateKeyFromSeed(seedBytes)
	if err != nil {
		panic(err)
	}
	return pk
}
