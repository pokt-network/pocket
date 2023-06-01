package utils

import "github.com/pokt-network/pocket/shared/mempool"

// Nonce is a type alias for uint64
type Nonce = uint64

// NonceDeduper is a type alias for the generic FIFO set with uint64 keys and values
type NonceDeduper = *mempool.GenericFIFOSet[Nonce, Nonce]

// NewNonceDeduper returns a new NonceDeduper with the given capacity
func NewNonceDeduper(capacity uint64) NonceDeduper {
	return mempool.NewGenericFIFOSet[Nonce, Nonce](int(capacity))
}
