package kvstore

import (
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

// type KVStoreConfig struct {
// 	opt := badger.DefaultOptions("").WithInMemory(true)
// }

type KVStore interface {
	// Create(name string, inMemory bool)

	// Start() error
	Stop() error

	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
}

type badgerKVStore struct {
	db *badger.DB

	name     string
	inMemory bool
}

// func Create() {
// 	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()
// }

func NewMemKVStore() KVStore {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	return badgerKVStore{db: db}

}

func NewKVStore(path string) (KVStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return badgerKVStore{db: db}, nil
}

func (store badgerKVStore) Get(key []byte) ([]byte, error) {
	// return store.db.Update(func(txn *badger.Txn) error {
	// 	return txn.Set(key, value)
	// }
	return nil, nil
}

func (store badgerKVStore) Put(key []byte, value []byte) error {
	// return store.db.Update(func(txn *badger.Txn) error {
	// 	return txn.Set(key, value)
	// }
	return nil
}

func (store badgerKVStore) Stop() error {
	return store.db.Close()
}
