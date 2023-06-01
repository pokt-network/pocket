package stores

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.Store = (*PrivateStore)(nil)

// PrivateStore does not need to be provable and as such simply wraps the KVStore interface
type PrivateStore struct {
	kvstore.KVStore
	storeKey string
	provable bool
}

// NewTestPrivateStore creates a new store for testing purposes using an in memory KVStore
func NewTestPrivateStore(storeKey string) modules.Store {
	db := kvstore.NewMemKVStore()
	return &PrivateStore{db, storeKey, false}
}

func (priv *PrivateStore) GetStoreKey() string {
	return priv.storeKey
}

func (priv *PrivateStore) IsProvable() bool {
	return priv.provable
}

// Get returns a value stored in the KVStore
func (priv *PrivateStore) Get(key []byte) ([]byte, error) {
	return priv.KVStore.Get(key)
}

// Set sets a value in the KVStore
func (priv *PrivateStore) Set(key, value []byte) error {
	return priv.KVStore.Set(key, value)
}

// Delete deletes a value from the KVStore
func (priv *PrivateStore) Delete(key []byte) error {
	return priv.KVStore.Delete(key)
}

// Stop closes the KVStore
func (priv *PrivateStore) Stop() error {
	return priv.KVStore.Stop()
}
