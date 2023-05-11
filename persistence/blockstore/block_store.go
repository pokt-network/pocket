package blockstore

//go:generate mockgen -source=$GOFILE -destination=../types/mocks/block_store_mock.go github.com/pokt-network/pocket/persistence/types BlockStore

import (
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
)

type BlockStore interface {
	kvstore.KVStore
}

var _ BlockStore = &blockStore{}

// BlockStore wraps a KVStore to provide atomic block storage.
// It provides the thin wrapper that manages the atomic state
// transitions for the application of a Unit of Work.
type blockStore struct {
	path string
	kv   kvstore.KVStore
}

// NewBlockStore initializes a new blockstore with the given path.
// * If "" is provided as the path, an in-memory store is used.
func NewBlockStore(path string) (*blockStore, error) {
	if path == "" {
		return &blockStore{
			path: "",
			kv:   kvstore.NewMemKVStore(),
		}, nil
	}
	kv, err := kvstore.NewKVStore(path)
	if err != nil {
		return nil, fmt.Errorf("failed to init blockstore kv: %w", err)
	}
	return &blockStore{
		path: path,
		kv:   kv,
	}, nil
}

// Set adds a block into the blockstore.
func (bs *blockStore) Set(k []byte, v []byte) error {
	return bs.kv.Set(k, v)
}

// Get returns a block at the provided height from the blockstore.
func (bs *blockStore) Get(key []byte) ([]byte, error) {
	return bs.kv.Get(key)
}

// ClearAll removes all blocks from the block store.
func (bs *blockStore) ClearAll() error {
	return bs.kv.ClearAll()
}

// Stop gracefully shuts down the blockstore.
func (bs *blockStore) Stop() error {
	return bs.kv.Stop()
}

// TODO
func (bs *blockStore) Delete(key []byte) error         { return nil }
func (bs *blockStore) Exists(key []byte) (bool, error) { return false, nil }
func (bs *blockStore) GetAll(prefixKey []byte, descending bool) (keys, values [][]byte, err error) {
	return nil, nil, nil
}
