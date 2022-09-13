package kvstore

import (
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

// CLEANUP: move this structure to a shared module
type KVStore interface {
	// Lifecycle methods
	Stop() error

	// Accessors
	// TODO: Add a proper iterator interface
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	// TODO: Add pagination for `GetAll`
	GetAll(prefixKey []byte, descending bool) ([][]byte, error)
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

// prefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		} else {
			end = end[:len(end)-1]
			if len(end) == 0 {
				end = nil
				break
			}
		}
	}
	return end
}
