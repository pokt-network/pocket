package kvstore

//go:generate mockgen -package=mock_types -destination=../types/mocks/kv_store_mock.go github.com/pokt-network/pocket/persistence/kvstore KVStore

import (
	"errors"
	"log"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/smt"
)

type KVStore interface {
	smt.MapStore // Get, Set, Delete

	// Lifecycle methods
	Stop() error

	// Accessors
	// TODO: Add a proper iterator interface
	// TODO: Add pagination for `GetAll`
	GetAll(prefixKey []byte, descending bool) (keys, values [][]byte, err error)
	Exists(key []byte) (bool, error)
	ClearAll() error
}

const (
	BadgerKeyNotFoundError = "Key not found"
)

var (
	_ KVStore      = &badgerKVStore{}
	_ smt.MapStore = &badgerKVStore{}
)

var (
	ErrKVStoreExists    = errors.New("kvstore already exists")
	ErrKVStoreNotExists = errors.New("kvstore does not exist")
)

type badgerKVStore struct {
	db *badger.DB
}

func NewKVStore(path string) (KVStore, error) {
	db, err := badger.Open(badgerOptions(path))
	if err != nil {
		return nil, err
	}
	return &badgerKVStore{db: db}, nil
}

func NewMemKVStore() KVStore {
	db, err := badger.Open(badgerOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	return &badgerKVStore{db: db}
}

func (store *badgerKVStore) Set(key, value []byte) error {
	return store.db.Update(func(tx *badger.Txn) error {
		return tx.Set(key, value)
	})
}

func (store *badgerKVStore) Get(key []byte) ([]byte, error) {
	var val []byte
	err := store.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return err
		}

		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (store *badgerKVStore) Delete(key []byte) error {
	return store.db.Update(func(tx *badger.Txn) error {
		return tx.Delete(key)
	})
}

func (store *badgerKVStore) GetAll(prefix []byte, descending bool) (keys, values [][]byte, err error) {
	err = store.db.View(func(tx *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = prefix
		opt.Reverse = descending
		if descending {
			prefix = prefixEndBytes(prefix)
		}
		it := tx.NewIterator(opt)
		defer it.Close()

		keys = make([][]byte, 0)
		values = make([][]byte, 0)

		for it.Seek(prefix); it.Valid(); it.Next() {
			item := it.Item()
			err = item.Value(func(v []byte) error {
				b := make([]byte, len(v))
				copy(b, v)
				keys = append(keys, item.Key())
				values = append(values, b)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return
}

func (store *badgerKVStore) Exists(key []byte) (bool, error) {
	val, err := store.Get(key)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

func (store *badgerKVStore) ClearAll() error {
	return store.db.DropAll()
}

func (store *badgerKVStore) Stop() error {
	return store.db.Close()
}

// PrefixEndBytes returns the end byteslice for a noninclusive range
// that would include all byte slices for which the input is the prefix
func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	if prefix[len(prefix)-1] == byte(255) {
		return prefixEndBytes(prefix[:len(prefix)-1])
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++
	return end
}

// TODO: Propagate persistence configurations to badger
func badgerOptions(path string) badger.Options {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // disable badger's logger since it's very noisy
	return opts
}
