package types

import (
	"sync"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
)

var _ mempool.TXMempool = &fIFOMempool{}

type fIFOMempool struct {
	g          *mempool.GenericFIFOSet[string, []byte]
	m          sync.Mutex
	size       uint32 // current number of transactions in the mempool
	txBytes    uint64 // current sum of all transactions' sizes (in bytes)
	maxTxBytes uint64 // maximum total size of all txs allowed in the mempool
}

// AddTx adds a tx to the mempool
func (t *fIFOMempool) AddTx(tx []byte) error {
	if err := t.g.Push(tx); err != nil {
		return ErrDuplicateTransaction()
	}
	return nil
}

// Clear clears the mempool
func (t *fIFOMempool) Clear() {
	t.g.Clear()
}

// Contains checks if a tx is in the mempool by its hash
func (t *fIFOMempool) Contains(hash string) bool {
	return t.g.ContainsIndex(hash)
}

// IsEmpty checks if the mempool is empty
func (t *fIFOMempool) IsEmpty() bool {
	return t.g.IsEmpty()
}

// PopTx pops a tx from the mempool
func (t *fIFOMempool) PopTx() ([]byte, error) {
	popTx, err := t.g.Pop()
	return []byte(popTx), NewError(-1, err.Error())
}

// RemoveTx removes a tx from the mempool
func (t *fIFOMempool) RemoveTx(tx []byte) error {
	t.g.Remove(tx)
	return nil
}

// TxCount returns the number of txs in the mempool
func (t *fIFOMempool) TxCount() uint32 {
	t.m.Lock()
	defer t.m.Unlock()
	return uint32(t.size)
}

// TxsBytesTotal returns the total size, in bytes, of all txs in the mempool
func (t *fIFOMempool) TxsBytesTotal() uint64 {
	t.m.Lock()
	defer t.m.Unlock()
	return t.txBytes
}

func NewTxFIFOMempool(maxTransactionBytes uint64, maxTransactions uint32) *fIFOMempool {
	txFIFOMempool := &fIFOMempool{
		m:          sync.Mutex{},
		size:       0,
		txBytes:    0,
		maxTxBytes: maxTransactionBytes,
	}

	txFIFOMempool.g = mempool.NewGenericFIFOSet(
		int(maxTransactions),
		mempool.WithIndexerFn[string, []byte](func(txBz any) string {
			return crypto.GetHashStringFromBytes(txBz.([]byte))
		}),
		mempool.WithCustomIsOverflowingFn(func(g *mempool.GenericFIFOSet[string, []byte]) bool {
			txFIFOMempool.m.Lock()
			defer txFIFOMempool.m.Unlock()
			return txFIFOMempool.size >= maxTransactions || txFIFOMempool.txBytes >= txFIFOMempool.maxTxBytes
		}),
		mempool.WithOnCollision(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) error {
			return ErrDuplicateTransaction()
		}),
		mempool.WithOnAdd(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) {
			txFIFOMempool.m.Lock()
			defer txFIFOMempool.m.Unlock()
			txFIFOMempool.size++
			txFIFOMempool.txBytes += uint64(len(item))
		}),
		mempool.WithOnRemove(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) {
			txFIFOMempool.m.Lock()
			defer txFIFOMempool.m.Unlock()
			txFIFOMempool.size--
			txFIFOMempool.txBytes -= uint64(len(item))
		}),
	)

	return txFIFOMempool
}
