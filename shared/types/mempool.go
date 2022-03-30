package types

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"sync"

	"github.com/pokt-network/pocket/shared/crypto"
)

type Mempool interface {
	Contains(hash string) bool
	AddTransaction(tx []byte) Error
	DeleteTransaction(tx []byte) Error

	Clear()
	Size() int
	TxsBytes() int
	PopTransaction() (tx []byte, sizeInBytes int, err Error) // TODO(andrew): In the upcoming merge, remove `sizeInBytes` from return value
}

var _ Mempool = &FIFOMempool{}

type FIFOMempool struct {
	l                    sync.RWMutex
	hashMap              map[string]struct{}
	pool                 *list.List
	size                 int
	transactionBytes     int
	maxTransactionsBytes int
	maxTransactions      int
}

func NewMempool(maxTransactionBytes int, maxTransactions int) Mempool {
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
	hash := crypto.SHA3Hash(tx)
	hashString := hex.EncodeToString(hash)
	if _, ok := f.hashMap[hashString]; ok {
		return ErrDuplicateTransaction()
	}
	f.pool.PushBack(tx)
	f.hashMap[hashString] = struct{}{}
	f.size++
	f.transactionBytes += len(tx)
	for f.size >= f.maxTransactions || f.transactionBytes >= f.maxTransactionsBytes {
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

func (f *FIFOMempool) PopTransaction() ([]byte, int, Error) {
	tx, err := popTransaction(f)
	if err != nil {
		return nil, 0, err
	}
	size := len(tx)
	return tx, size, nil
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
	hash := crypto.SHA3Hash(txBz)
	hashString := hex.EncodeToString(hash)
	delete(f.hashMap, hashString)
	f.size--
	f.transactionBytes -= txBzLen
	return txBz, nil
}

func popTransaction(f *FIFOMempool) ([]byte, Error) {
	return removeTransaction(f, f.pool.Front())
}
