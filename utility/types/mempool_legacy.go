package types

import (
	"bytes"
	"container/list"
	"sync"

	"github.com/pokt-network/pocket/shared/crypto"
)

var _ Mempool = &legacyFIFOMempool{}

type legacyFIFOMempool struct {
	l          sync.RWMutex
	txSet      map[string]struct{} // a set used to determine the existence of a tx by its hash
	txQueue    *list.List          // doubly linked list of transactions
	size       uint32              // current number of transactions in the mempool
	txBytes    uint64              // current sum of all transactions' sizes (in bytes)
	maxTxBytes uint64              // maximum total size of all txs allowed in the mempool
	maxTxs     uint32              // maximum number of transactions allowed in the mempool
}

func NewMempool(maxTransactionBytes uint64, maxTransactions uint32) Mempool {
	return &legacyFIFOMempool{
		l:          sync.RWMutex{},
		txSet:      make(map[string]struct{}),
		txQueue:    list.New(),
		size:       0,
		txBytes:    0,
		maxTxBytes: maxTransactionBytes,
		maxTxs:     maxTransactions,
	}
}

func (f *legacyFIFOMempool) AddTransaction(tx []byte) Error {
	f.l.Lock()
	defer f.l.Unlock()

	// Check if present
	hashString := crypto.GetHashStringFromBytes(tx)
	if _, ok := f.txSet[hashString]; ok {
		return ErrDuplicateTransaction()
	}

	// Insert the tx into the mempool
	f.txQueue.PushBack(tx)
	f.txSet[hashString] = struct{}{}
	f.size++
	f.txBytes += uint64(len(tx))

	// TODO: Rather than inserting the tx and than popping - we should just insert the tx only after validation
	// Remove the tx if it exceeds the configs
	for f.size >= f.maxTxs || f.txBytes >= f.maxTxBytes {
		_, err := f.popTransaction()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *legacyFIFOMempool) Contains(hash string) bool {
	f.l.RLock()
	defer f.l.RUnlock()

	if _, has := f.txSet[hash]; has {
		return true
	}
	return false
}

func (f *legacyFIFOMempool) RemoveTransaction(tx []byte) Error {
	f.l.Lock()
	defer f.l.Unlock()

	var toRemove *list.Element
	for e := f.txQueue.Front(); e.Next() != nil; {
		if bytes.Equal(tx, e.Value.([]byte)) {
			toRemove = e
			break
		}
	}
	if toRemove != nil {
		_, err := f.removeTransaction(toRemove)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *legacyFIFOMempool) PopTransaction() ([]byte, Error) {
	f.l.RLock()
	defer f.l.RUnlock()

	tx, err := f.popTransaction()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (f *legacyFIFOMempool) Clear() {
	f.l.Lock()
	defer f.l.Unlock()

	f.txSet = make(map[string]struct{})
	f.txQueue = list.New()
	f.size = 0
	f.txBytes = 0
}

func (f *legacyFIFOMempool) Size() uint32 {
	f.l.RLock()
	defer f.l.RUnlock()

	return f.size
}

func (f *legacyFIFOMempool) IsEmpty() bool {
	f.l.RLock()
	defer f.l.RUnlock()

	return f.size == 0
}

func (f *legacyFIFOMempool) TxsBytes() uint64 {
	f.l.RLock()
	defer f.l.RUnlock()

	return f.txBytes
}

func (f *legacyFIFOMempool) popTransaction() ([]byte, Error) {
	return f.removeTransaction(f.txQueue.Front())
}

func (f *legacyFIFOMempool) removeTransaction(e *list.Element) ([]byte, Error) {
	if f.size == 0 {
		return nil, nil
	}

	txBz := e.Value.([]byte)
	txBzLen := uint64(len(txBz))
	f.txQueue.Remove(e)

	hashString := crypto.GetHashStringFromBytes(txBz)
	delete(f.txSet, hashString)

	f.size--
	f.txBytes -= txBzLen

	return txBz, nil
}
