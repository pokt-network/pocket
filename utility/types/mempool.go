package types

import (
	"sync"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
)

type Mempool interface {
	Contains(hash string) bool
	AddTransaction(tx []byte) Error
	RemoveTransaction(tx []byte) Error

	Clear()
	Size() uint32 // Returns the number of transactions stored in the mempool
	IsEmpty() bool
	TxsBytes() uint64 // Returns the total sum of all transactions' sizes (in bytes) stored in the mempool
	PopTransaction() (tx []byte, err Error)
}

var _ Mempool = &fIFOMempool{}

type fIFOMempool struct {
	g          *mempool.GenericFIFOSet[string, []byte]
	m          sync.Mutex
	size       uint32 // current number of transactions in the mempool
	txBytes    uint64 // current sum of all transactions' sizes (in bytes)
	maxTxBytes uint64 // maximum total size of all txs allowed in the mempool
}

// AddTransaction implements Mempool
func (t *fIFOMempool) AddTransaction(tx []byte) Error {
	if err := t.g.Push(tx); err != nil {
		return ErrDuplicateTransaction()
	}
	return nil
}

func (t *fIFOMempool) Clear() {
	t.g.Clear()
}

func (t *fIFOMempool) Contains(hash string) bool {
	return t.g.ContainsIndex(hash)
}

func (t *fIFOMempool) IsEmpty() bool {
	return t.g.IsEmpty()
}

func (t *fIFOMempool) PopTransaction() ([]byte, Error) {
	popTx, err := t.g.Pop()
	return []byte(popTx), NewError(-1, err.Error()) // TODO: prettier
}

func (t *fIFOMempool) RemoveTransaction(tx []byte) Error {
	t.g.Remove(tx)
	return nil
}

func (t *fIFOMempool) Size() uint32 {
	t.m.Lock()
	defer t.m.Unlock()
	return uint32(t.size)
}

func (t *fIFOMempool) TxsBytes() uint64 {
	t.m.Lock()
	defer t.m.Unlock()
	return t.txBytes
}

func NewTxFIFOMempool(maxTransactionBytes uint64, maxTransactions uint32) *fIFOMempool {
	txFifoMempool := &fIFOMempool{
		m:          sync.Mutex{},
		size:       0,
		txBytes:    0,
		maxTxBytes: maxTransactionBytes,
	}

	txFifoMempool.g = mempool.NewGenericFIFOSet(
		int(maxTransactions),
		mempool.WithIndexerFn[string, []byte](func(txBz any) string {
			return crypto.GetHashStringFromBytes(txBz.([]byte))
		}),
		mempool.WithCustomIsOverflowingFn(func(g *mempool.GenericFIFOSet[string, []byte]) bool {
			txFifoMempool.m.Lock()
			defer txFifoMempool.m.Unlock()
			return txFifoMempool.size >= maxTransactions || txFifoMempool.txBytes >= txFifoMempool.maxTxBytes
		}),
		mempool.WithOnCollision(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) error {
			return ErrDuplicateTransaction()
		}),
		mempool.WithOnAdd(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) {
			txFifoMempool.m.Lock()
			defer txFifoMempool.m.Unlock()
			txFifoMempool.size++
			txFifoMempool.txBytes += uint64(len(item))
		}),
		mempool.WithOnRemove(func(item []byte, g *mempool.GenericFIFOSet[string, []byte]) {
			txFifoMempool.m.Lock()
			defer txFifoMempool.m.Unlock()
			txFifoMempool.size--
			txFifoMempool.txBytes -= uint64(len(item))
		}),
	)

	return txFifoMempool
}
