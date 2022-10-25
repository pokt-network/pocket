package kvstore

import (
	"errors"
	"log"

	"github.com/celestiaorg/smt"
	badger "github.com/dgraph-io/badger/v3"
)

type KVStore interface {
	smt.MapStore

	// Lifecycle methods
	Stop() error

	// Accessors
	// TODO: Add a proper iterator interface
	// TODO: Add pagination for `GetAll`
	GetAll(prefixKey []byte, descending bool) ([][]byte, error)

	Exists(key []byte) (bool, error)
	ClearAll() error
}

const (
	BadgerKeyNotFoundError = "Key not found"
)

var _ KVStore = &badgerKVStore{}
var _ smt.MapStore = &badgerKVStore{}

var (
	ErrKVStoreExists    = errors.New("kvstore already exists")
	ErrKVStoreNotExists = errors.New("kvstore does not exist")
)

type badgerKVStore struct {
	db *badger.DB
}

func NewKVStore(path string) (KVStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &badgerKVStore{db: db}, nil
}

func NewMemKVStore() KVStore {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	return &badgerKVStore{db: db}
}

func (store *badgerKVStore) Set(key, value []byte) error {
	tx := store.db.NewTransaction(true)
	defer tx.Discard()

	err := tx.Set(key, value)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (store *badgerKVStore) Get(key []byte) ([]byte, error) {
	tx := store.db.NewTransaction(false)
	defer tx.Discard()

	item, err := tx.Get(key)
	if err != nil {
		return nil, err
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return value, nil
}

func (store *badgerKVStore) Delete(key []byte) error {
	log.Fatalf("badgerKVStore.Delete not implemented yet")
	return nil
}

func (store *badgerKVStore) GetAll(prefix []byte, descending bool) (values [][]byte, err error) {
	// INVESTIGATE: research `badger.views` for further improvements and optimizations
	txn := store.db.NewTransaction(false)
	defer txn.Discard()

	opt := badger.DefaultIteratorOptions
	opt.Prefix = prefix
	opt.Reverse = descending
	if descending {
		prefix = prefixEndBytes(prefix)
	}
	it := txn.NewIterator(opt)
	defer it.Close()

	for it.Seek(prefix); it.Valid(); it.Next() {
		item := it.Item()
		err = item.Value(func(v []byte) error {
			b := make([]byte, len(v))
			copy(b, v)
			values = append(values, b)
			return nil
		})
		if err != nil {
			return
		}
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
