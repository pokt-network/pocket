package crypto

import (
	crand "crypto/rand"
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
	return bigNonce.Uint64()
}
