package stores

import (
	"crypto/sha256"
	"sync"

	"github.com/pokt-network/pocket/ibc/host"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.StoreManager = (*Stores)(nil)

type Stores struct {
	m      sync.Mutex
	stores map[string]modules.Store
}

// NewStoreManager creates a new store manager instance with an empty map of stores
func NewStoreManager() modules.StoreManager {
	return &Stores{
		stores: make(map[string]modules.Store),
	}
}

// NewStore creates a new store using a persistent KVStore at the path provided
func NewStore(storeKey, storePath string, provable bool) (modules.Store, error) {
	db, err := kvstore.NewKVStore(storePath)
	if err != nil {
		return nil, coreTypes.ErrStoreCreation(err)
	}
	if !provable {
		return &PrivateStore{db, storeKey, false}, nil
	}
	tree := smt.NewSparseMerkleTree(db, sha256.New(), noValueHasher)
	return &ProvableStore{db, tree, storeKey, true}, nil
}

// InitialiseStore populates a store with the data provided by iterating through the map,
// applying the store prefix to the paths and setting the value in the store to this key
func InitialiseStore(store modules.Store, data map[string][]byte) error {
	for path, value := range data {
		prefix := &coreTypes.CommitmentPrefix{Prefix: []byte(store.GetStoreKey())}
		key := host.ApplyPrefix(prefix, path).GetPath()
		if err := store.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetStore returns a store instance from the store manager
func (s *Stores) GetStore(storeKey string) (modules.Store, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.stores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	return store, nil
}

// GetProvableStore retrieves a ProvableStore instance from the StoreManager
func (s *Stores) GetProvableStore(storeKey string) (modules.ProvableStore, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.stores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	if !store.IsProvable() {
		return nil, coreTypes.ErrStoreNotProvable(storeKey)
	}
	return store.(modules.ProvableStore), nil
}

// AddStore adds a Store instance to the StoreManager
func (s *Stores) AddStore(store modules.Store) error {
	s.m.Lock()
	defer s.m.Unlock()
	storeKey := store.GetStoreKey()
	if _, ok := s.stores[storeKey]; ok {
		return coreTypes.ErrStoreAlreadyExists(storeKey)
	}
	s.stores[storeKey] = store
	return nil
}

// RemoveStore removes a Store instance from the StoreManager
func (s *Stores) RemoveStore(storeKey string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[storeKey]; !ok {
		return coreTypes.ErrStoreNotFound(storeKey)
	}
	delete(s.stores, storeKey)
	return nil
}
