package raintree

import "github.com/pokt-network/pocket/shared/mempool"

func NewNonceDeduper(mempoolMaxNonces uint64) *mempool.GenericFIFOSet[uint64, uint64] {
	return mempool.NewGenericFIFOSet[uint64, uint64](int(mempoolMaxNonces))
}
