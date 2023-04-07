package kvstore

//go:generate mockgen -source=$GOFILE -destination=../types/mocks/block_store_mock.go github.com/pokt-network/pocket/persistence/types KVStore

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/celestiaorg/smt"
	badger "github.com/dgraph-io/badger/v3"
)

type KVStore interface {
	smt.MapStore // Get, Set, Delete
	BackupableKVStore

	// Lifecycle methods
	Stop() error

	// Accessors
	// TODO: Add a proper iterator interface
	// TODO: Add pagination for `GetAll`
	GetAll(prefixKey []byte, descending bool) (keys, values [][]byte, err error)
	Exists(key []byte) (bool, error)
	ClearAll() error
}

type BackupableKVStore interface {
	GetName() string
	Backup(w io.Writer, since uint64) (uint64, error)
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
	db   *badger.DB
	name string
}

func NewKVStore(path string) (KVStore, error) {
	db, err := badger.Open(badgerOptions(path))
	if err != nil {
		return nil, err
	}

	name, err := extractNameFromPath(path)
	if err != nil {
		return nil, err
	}

	return &badgerKVStore{
		db:   db,
		name: name,
	}, nil
}

func NewMemKVStore(name string) KVStore {
	db, err := badger.Open(badgerOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	return &badgerKVStore{
		db:   db,
		name: name,
	}
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
	tx := store.db.NewTransaction(true)
	defer tx.Discard()

	return tx.Delete(key)
}

func (store *badgerKVStore) GetAll(prefix []byte, descending bool) (keys, values [][]byte, err error) {
	// INVESTIGATE: research `badger.views` for further improvements and optimizations
	// Reference https://pkg.go.dev/github.com/dgraph-io/badger#readme-prefix-scans
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

// Backup writes a backup of the database to the given writer.
func (store *badgerKVStore) Backup(w io.Writer, since uint64) (uint64, error) {
	return store.db.Backup(w, since)
}

func (store *badgerKVStore) GetName() string {
	return store.name
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

func extractNameFromPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	parts := strings.Split(path, "/")
	nodesPart := parts[len(parts)-1]
	name := strings.TrimSuffix(nodesPart, "_nodes")

	if name == "" {
		return "", fmt.Errorf("invalid path format, name not found in %s", path)
	}

	return name, nil
}
