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
}

func (priv *PrivateStore) GetStoreKey() string {
	return priv.storeKey
}

func (priv *PrivateStore) Get(key []byte) ([]byte, error) {
	return priv.KVStore.Get(key)
}

func (priv *PrivateStore) Set(key, value []byte) error {
	return priv.KVStore.Set(key, value)
}

func (priv *PrivateStore) Delete(key []byte) error {
	return priv.KVStore.Delete(key)
}
