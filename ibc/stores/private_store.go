package stores

//go:generate mockgen -package=mock_types -destination=../types/mocks/private_store_mock.go github.com/pokt-network/pocket/ibc/stores PrivateStore

import (
	"github.com/pokt-network/pocket/ibc/host"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.PrivateStore = (*PrivateStore)(nil)

// PrivateStore does not need to be provable and as such simply wraps the KVStore interface
type PrivateStore struct {
	kvstore.KVStore
	storeKey string
}

// NewTestPrivateStore creates a new store for testing purposes using an in memory KVStore
func NewTestPrivateStore(storeKey string) (modules.PrivateStore, error) {
	db := kvstore.NewMemKVStore()
	return &PrivateStore{db, storeKey}, nil
}

// NewPrivateStore creates a new store using a persistent KVStore at the path provided
func NewPrivateStore(storeKey, storePath string) (modules.PrivateStore, error) {
	db, err := kvstore.NewKVStore(storePath)
	if err != nil {
		return nil, coreTypes.ErrStoreCreation(err)
	}
	return &PrivateStore{db, storeKey}, nil
}

// InitialiseStore populates a store with the data provided by iterating through the map,
// applying the store prefix to the paths and setting the value in the store to this key
func InitialiseStore(store modules.PrivateStore, data map[string][]byte) error {
	for path, value := range data {
		prefix := &coreTypes.CommitmentPrefix{Prefix: []byte(store.GetStoreKey())}
		key := host.ApplyPrefix(prefix, path).GetPath()
		if err := store.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (priv *PrivateStore) GetStoreKey() string {
	return priv.storeKey
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
