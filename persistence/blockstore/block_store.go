package blockstore

//go:generate mockgen -source=$GOFILE -destination=../types/mocks/blockstore/block_store_mock.go github.com/pokt-network/pocket/persistence/types BlockStore

import (
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
)

// BlockStore is a key-value store that maps block heights to serialized
// block structures.
// * It manages the atomic state transitions for applying a Unit of Work.
type BlockStore interface {
	kvstore.KVStore
}

// Enforce blockStore to fulfill BlockStore
var _ BlockStore = &blockStore{}

// blockStore wraps a KVStore to provide atomic commits
// and implements `BlockStore`
// * It's responsible for the lifecycle of savepoints and
// rollbacks with its underlying KVStore.
type blockStore struct {
	kv kvstore.KVStore
}

// NewBlockStore initializes a new blockstore with the given path.
// * If "" is provided as the path, an in-memory store is used.
func NewBlockStore(path string) (BlockStore, error) {
	if path == "" {
		return &blockStore{
			kv: kvstore.NewMemKVStore(),
		}, nil
	}
	kv, err := kvstore.NewKVStore(path)
	if err != nil {
		return nil, fmt.Errorf("failed to init blockstore kv: %w", err)
	}
	return &blockStore{
		kv: kv,
	}, nil
}

// Set adds a block into the blockstore.
func (bs *blockStore) Set(k, v []byte) error {
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

func (bs *blockStore) Delete(key []byte) error         { return bs.kv.Delete(key) }
func (bs *blockStore) Exists(key []byte) (bool, error) { return bs.kv.Exists(key) }
func (bs *blockStore) GetAll(prefixKey []byte, descending bool) (keys, values [][]byte, err error) {
	return bs.kv.GetAll(prefixKey, descending)
}
