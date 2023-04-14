package types

import (
	"sync"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
)

var _ mempool.TXMempool = &txFIFOMempool{}

type txFIFOMempool struct {
	g                *mempool.GenericFIFOSet[string, []byte]
	m                sync.Mutex
	txCount          uint32 // current number of transactions in the mempool
	txsBytesTotal    uint64 // current sum of all transactions' sizes (in bytes)
	maxTxsBytesTotal uint64 // maximum total size of all txs allowed in the mempool
}

// AddTx adds a tx to the mempool
func (t *txFIFOMempool) AddTx(tx []byte) coreTypes.Error {
	if err := t.g.Push(tx); err != nil {
		return coreTypes.ErrDuplicateTransaction()
	}
	return nil
}

// Clear clears the mempool
func (t *txFIFOMempool) Clear() {
	t.g.Clear()
	resetCounters(t)
}

// resetCounters resets txCount and txsBytesTotal to 0
func resetCounters(t *txFIFOMempool) {
	t.m.Lock()
	defer t.m.Unlock()
	t.txCount = 0
	t.txsBytesTotal = 0
}

// Contains checks if a tx is in the mempool by its hash
func (t *txFIFOMempool) Contains(hash string) bool {
	return t.g.ContainsIndex(hash)
}

// IsEmpty checks if the mempool is empty
func (t *txFIFOMempool) IsEmpty() bool {
	return t.g.IsEmpty()
}

// PopTx pops a tx from the mempool
func (t *txFIFOMempool) PopTx() ([]byte, error) {
	return t.g.Pop()
}

// RemoveTx removes a tx from the mempool
func (t *txFIFOMempool) RemoveTx(tx []byte) error {
	t.g.Remove(tx)
	return nil
}

// TxCount returns the number of txs in the mempool
func (t *txFIFOMempool) TxCount() uint32 {
	t.m.Lock()
	defer t.m.Unlock()
	return uint32(t.txCount)
}

// TxsBytesTotal returns the total size, in bytes, of all txs in the mempool
func (t *txFIFOMempool) TxsBytesTotal() uint64 {
	t.m.Lock()
	defer t.m.Unlock()
	return t.txsBytesTotal
}

func NewTxFIFOMempool(maxTxsBytesTotal uint64, maxTxs uint32) *txFIFOMempool {
	txFIFOMempool := &txFIFOMempool{
		m:                sync.Mutex{},
		txCount:          0,
		txsBytesTotal:    0,
		maxTxsBytesTotal: maxTxsBytesTotal,
	}

	txFIFOMempool.g = mempool.NewGenericFIFOSet(
		int(maxTxs),
		mempool.WithIndexerFn[string, []byte](func(txBz any) string {
			return crypto.GetHashStringFromBytes(txBz.([]byte))
		}),
		mempool.WithCustomIsOverflowingFn(func(g *mempool.GenericFIFOSet[string, []byte]) bool {
			txFIFOMempool.m.Lock()
			defer txFIFOMempool.m.Unlock()
			return txFIFOMempool.txCount > maxTxs || txFIFOMempool.txsBytesTotal > txFIFOMempool.maxTxsBytesTotal
		}),
		mempool.WithOnCollision(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) error {
			return coreTypes.ErrDuplicateTransaction()
		}),
		mempool.WithOnAdd(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) {
			txFIFOMempool.m.Lock()
			defer txFIFOMempool.m.Unlock()
			txFIFOMempool.txCount++
			txFIFOMempool.txsBytesTotal += uint64(len(item))
		}),
		mempool.WithOnRemove(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) {
			txFIFOMempool.m.Lock()
			defer txFIFOMempool.m.Unlock()
			txFIFOMempool.txCount--
			txFIFOMempool.txsBytesTotal -= uint64(len(item))
		}),
	)

	return txFIFOMempool
}
