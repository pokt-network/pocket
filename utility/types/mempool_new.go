package types

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/mempool"
)

var _ Mempool = &TxFIFOMempool{}

type TxFIFOMempool struct {
	g          *mempool.GenericFIFOSet[string, string]
	m          sync.Mutex
	size       uint32 // current number of transactions in the mempool
	txBytes    uint64 // current sum of all transactions' sizes (in bytes)
	maxTxBytes uint64 // maximum total size of all txs allowed in the mempool
}

// AddTransaction implements Mempool
func (t *TxFIFOMempool) AddTransaction(tx []byte) Error {
	if err := t.g.Push(string(tx)); err != nil {
		return ErrDuplicateTransaction()
	}
	return nil
}

// Clear implements Mempool
func (t *TxFIFOMempool) Clear() {
	t.g.Clear()
}

// Contains implements Mempool
func (t *TxFIFOMempool) Contains(hash string) bool {
	return t.g.Contains(hash)
}

// IsEmpty implements Mempool
func (t *TxFIFOMempool) IsEmpty() bool {
	return t.g.IsEmpty()
}

// PopTransaction implements Mempool
func (t *TxFIFOMempool) PopTransaction() ([]byte, Error) {
	popTx, err := t.g.Pop()
	return []byte(popTx), NewError(-1, err.Error()) // TODO: prettier
}

// RemoveTransaction implements Mempool
func (t *TxFIFOMempool) RemoveTransaction(tx []byte) Error {
	t.g.Remove(crypto.GetHashStringFromBytes(tx))
	return nil
}

// Size implements Mempool
func (t *TxFIFOMempool) Size() uint32 {
	t.m.Lock()
	defer t.m.Unlock()
	return uint32(t.size)
}

// TxsBytes implements Mempool
func (t *TxFIFOMempool) TxsBytes() uint64 {
	t.m.Lock()
	defer t.m.Unlock()
	return t.txBytes
}

func NewTxFIFOMempool(maxTransactionBytes uint64, maxTransactions uint32) *TxFIFOMempool {
	txFifoMempool := &TxFIFOMempool{
		m:          sync.Mutex{},
		size:       0,
		txBytes:    0,
		maxTxBytes: maxTransactionBytes,
	}

	txFifoMempool.g = mempool.NewGenericFIFOSet(
		int(maxTransactions),
		mempool.WithIndexerFn[string, string](func(txBz any) string {
			return crypto.GetHashStringFromBytes(txBz.([]byte))
		}),
		mempool.WithCustomIsOverflowingFn(func(g *mempool.GenericFIFOSet[string, string]) bool {
			txFifoMempool.m.Lock()
			defer txFifoMempool.m.Unlock()
			return txFifoMempool.size >= maxTransactions || txFifoMempool.txBytes >= txFifoMempool.maxTxBytes
		}),
		mempool.WithOnCollision(func(item string, g *mempool.GenericFIFOSet[string, string]) error {
			//return ErrDuplicateTransaction()
			return fmt.Errorf("duplicate transaction") // TODO: replace with ErrDuplicateTransaction
		}),
		mempool.WithOnAdd(func(item string, g *mempool.GenericFIFOSet[string, string]) {
			txFifoMempool.m.Lock()
			defer txFifoMempool.m.Unlock()
			txFifoMempool.size++
			txFifoMempool.txBytes += uint64(len(item))
		}),
		mempool.WithOnRemove(func(item string, g *mempool.GenericFIFOSet[string, string]) {
			txFifoMempool.m.Lock()
			defer txFifoMempool.m.Unlock()
			txFifoMempool.size--
			txFifoMempool.txBytes -= uint64(len(item))
		}),
	)

	return txFifoMempool
}
