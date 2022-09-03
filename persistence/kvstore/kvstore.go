package kvstore

import (
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

// TODO: move to shared
type KVStore interface {
	// Lifecycle methods
	Stop() error

	// Accessors
	// TODO (Team) need proper iterator interface, can't live on this interface without one
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	GetAll(prefixKey []byte, descending bool) ([][]byte, error) // TODO Pagination for GetAll()
	Exists(key []byte) (bool, error)
	ClearAll() error
}

var _ KVStore = &badgerKVStore{}

type badgerKVStore struct {
	db *badger.DB
}

func NewKVStore(path string) (KVStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return badgerKVStore{db: db}, nil
}

func NewMemKVStore() KVStore {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	return badgerKVStore{db: db}
}

func (store badgerKVStore) Put(key []byte, value []byte) error {
	txn := store.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key, value)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (store badgerKVStore) Get(key []byte) ([]byte, error) {
	txn := store.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return value, nil
}

func (store badgerKVStore) GetAll(prefix []byte, descending bool) (values [][]byte, err error) {
	// TODO (INVESTIGATE) research badger.views
	txn := store.db.NewTransaction(false)
	defer txn.Discard()

	opt := badger.DefaultIteratorOptions
	opt.Prefix = prefix
	opt.Reverse = !descending
	it := txn.NewIterator(opt)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		err = item.Value(func(v []byte) error {
			values = append(values, v)
			return nil
		})
		if err != nil {
			return
		}
	}
	return
}

func (store badgerKVStore) Exists(key []byte) (bool, error) {
	val, err := store.Get(key)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

func (store badgerKVStore) ClearAll() error {
	return store.db.DropAll()
}

func (store badgerKVStore) Stop() error {
	return store.db.Close()
}
