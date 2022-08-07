package kvstore

import (
	"errors"
	"log"

	"github.com/celestiaorg/smt"
	badger "github.com/dgraph-io/badger/v3"
)

// TODO_IN_THIS_COMMIT: We might be able to remove the `KVStore` interface altogether if we end up using smt.MapStore
type KVStore interface {
	// Lifecycle methods
	Stop() error

	Exists(key []byte) (bool, error)
	ClearAll() error

	// Same interface as in `smt.MapStore``
	Set(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
}

var _ KVStore = &badgerKVStore{}
var _ smt.MapStore = &badgerKVStore{}

var (
	ErrKVStoreExists    = errors.New("kvstore already exists")
	ErrKVStoreNotExists = errors.New("kvstore does not exist")
)

type badgerKVStore struct {
	db *badger.DB
}

// REFACTOR: Loads or creates a badgerDb at `path`. This may potentially need to be refactored
// into `NewKVStore` and `LoadKVStore` depending on how state sync evolves by leveraging `os.Stat`
// on the file path.
func OpenKVStore(path string) (KVStore, error) {
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

func (store badgerKVStore) Set(key []byte, value []byte) error {
	txn := store.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key, value)
	if err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
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

func (store badgerKVStore) Delete(key []byte) error {
	log.Fatalf("badgerKVStore.Delete not implemented yet")
	return nil
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
