package types

import (
	"bytes"
	"container/list"
	"sync"

	"github.com/pokt-network/pocket/shared/crypto"
)

type Mempool interface {
	Contains(hash string) bool
	AddTransaction(tx []byte) Error
	DeleteTransaction(tx []byte) Error

	Clear()
	Size() int // TODO: Add IsEmpty() function
	TxsBytes() int
	PopTransaction() (tx []byte, err Error)
}

var _ Mempool = &FIFOMempool{}

type FIFOMempool struct {
	l                    sync.RWMutex
	hashMap              map[string]struct{}
	pool                 *list.List
	size                 int
	transactionBytes     int
	maxTransactionsBytes uint64
	maxTransactions      uint32
}

func NewMempool(maxTransactionBytes uint64, maxTransactions uint32) Mempool {
	return &FIFOMempool{
		l:                    sync.RWMutex{},
		hashMap:              make(map[string]struct{}),
		pool:                 list.New(),
		size:                 0,
		transactionBytes:     0,
		maxTransactionsBytes: maxTransactionBytes,
		maxTransactions:      maxTransactions,
	}
}

func (f *FIFOMempool) AddTransaction(tx []byte) Error {
	f.l.Lock()
	defer f.l.Unlock()
	hashString := crypto.GetHashStringFromBytes(tx)
	if _, ok := f.hashMap[hashString]; ok {
		return ErrDuplicateTransaction()
	}
	f.pool.PushBack(tx)
	f.hashMap[hashString] = struct{}{}
	f.size++
	f.transactionBytes += len(tx)
	for uint32(f.size) >= f.maxTransactions || uint64(f.transactionBytes) >= f.maxTransactionsBytes {
		_, err := popTransaction(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FIFOMempool) Contains(hash string) bool {
	f.l.RLock()
	defer f.l.RUnlock()
	if _, has := f.hashMap[hash]; has {
		return true
	}
	return false
}

func (f *FIFOMempool) DeleteTransaction(tx []byte) Error {
	f.l.Lock()
	defer f.l.Unlock()
	var toRemove *list.Element
	for e := f.pool.Front(); e.Next() != nil; {
		if bytes.Equal(tx, e.Value.([]byte)) {
			toRemove = e
			break
		}
	}
	if toRemove != nil {
		_, err := removeTransaction(f, toRemove)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FIFOMempool) PopTransaction() ([]byte, Error) {
	tx, err := popTransaction(f)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (f *FIFOMempool) Clear() {
	f.l.Lock()
	defer f.l.Unlock()
	f.pool = list.New()
	f.hashMap = make(map[string]struct{})
	f.size = 0
	f.transactionBytes = 0
}

func (f *FIFOMempool) Size() int {
	f.l.RLock()
	defer f.l.RUnlock()
	return f.size
}

func (f *FIFOMempool) TxsBytes() int {
	f.l.RLock()
	defer f.l.RUnlock()
	return f.transactionBytes
}

func removeTransaction(f *FIFOMempool, e *list.Element) ([]byte, Error) {
	if f.size == 0 {
		return nil, nil
	}
	txBz := e.Value.([]byte)
	txBzLen := len(txBz)
	f.pool.Remove(e)
	hashString := crypto.GetHashStringFromBytes(txBz)
	delete(f.hashMap, hashString)
	f.size--
	f.transactionBytes -= txBzLen
	return txBz, nil
}

func popTransaction(f *FIFOMempool) ([]byte, Error) {
	return removeTransaction(f, f.pool.Front())
}
