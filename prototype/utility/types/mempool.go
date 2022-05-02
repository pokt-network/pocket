package types

import (
	"container/list"
	"sync"
)

type Mempool interface {
	Contains(hash string) bool
	AddTransaction(tx *Transaction) Error
	DeleteTransaction(tx *Transaction) Error
	//Reap(maxTransactions int, maxTransactionBytes int) ([]*Transaction, error)
	Flush()
	Size() int
	TxsBytes() int
	PopTransaction() (txBytes []byte, tx *Transaction, sizeInBytes int, err Error)
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

func (f *FIFOMempool) AddTransaction(tx *Transaction) Error {
	f.l.Lock()
	defer f.l.Unlock()
	hash, err := tx.Hash()
	if err != nil {
		return err
	}
	if _, ok := f.hashMap[hash]; ok {
		return ErrDuplicateTransaction()
	}
	f.pool.PushBack(tx)
	f.hashMap[hash] = struct{}{}
	bz, err := tx.Bytes()
	if err != nil {
		return err
	}
	f.size++
	f.transactionBytes += len(bz)
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

func (f *FIFOMempool) DeleteTransaction(tx *Transaction) Error {
	f.l.Lock()
	defer f.l.Unlock()
	var toRemove *list.Element
	for e := f.pool.Front(); e.Next() != nil; {
		if tx.Equals(e.Value.(*Transaction)) {
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

func (f *FIFOMempool) PopTransaction() (txBytes []byte, tx *Transaction, size int, err Error) {
	tx, err = popTransaction(f)
	if err != nil {
		return
	}
	txBytes, err = tx.Bytes()
	if err != nil {
		return
	}
	size = len(txBytes)
	return
}

func (f *FIFOMempool) Reap(maxTransactions, maxTransactionBytes int) (t []*Transaction, err Error) {
	// simple fifo for now
	f.l.RLock()
	defer f.l.RUnlock()
	e := f.pool.Front()
	size := 0
	txBytes := 0
	for size >= maxTransactions || txBytes >= maxTransactionBytes && e.Next() != nil {
		tx := e.Value.(*Transaction)
		t = append(t, tx)
		size++
		bz, err := tx.Bytes()
		if err != nil {
			return nil, err
		}
		txBytes += len(bz)
	}
	return
}

func (f *FIFOMempool) Flush() {
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

func removeTransaction(f *FIFOMempool, e *list.Element) (*Transaction, Error) {
	if f.size == 0 {
		return nil, nil
	}
	t := e.Value.(*Transaction)
	bz, err := t.Bytes()
	if err != nil {
		return nil, err
	}
	tBz := len(bz)
	f.pool.Remove(e)
	h, err := t.Hash()
	if err != nil {
		return nil, err
	}
	delete(f.hashMap, h)
	f.size--
	f.transactionBytes -= tBz
	return t, nil
}

func popTransaction(f *FIFOMempool) (*Transaction, Error) {
	return removeTransaction(f, f.pool.Front())
}
