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
	m         sync.Mutex
	storesDir string
	stores    map[string]modules.Store
}

// NewStoreManager creates a new store manager instance with an empty map of stores
func NewStoreManager(storesDirPath string) modules.StoreManager {
	return &Stores{
		storesDir: storesDirPath,
		stores:    make(map[string]modules.Store),
	}
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
	provable, ok := store.(modules.ProvableStore)
	if !ok || !store.IsProvable() {
		return nil, coreTypes.ErrStoreNotProvable(storeKey)
	}
	return provable, nil
}

// AddStore adds a new Store instance to the StoreManager
func (s *Stores) AddStore(storeKey string, provable bool) (modules.Store, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[storeKey]; ok {
		return nil, coreTypes.ErrStoreAlreadyExists(storeKey)
	}
	var store modules.Store
	db, err := kvstore.NewKVStore(s.storesDir + "/" + storeKey)
	if err != nil {
		return nil, coreTypes.ErrStoreCreation(err)
	}
	if !provable {
		store = &PrivateStore{db, storeKey, false}
	} else {
		tree := smt.NewSparseMerkleTree(db, sha256.New(), noValueHasher)
		store = &ProvableStore{db, tree, storeKey, true}
	}
	s.stores[storeKey] = store
	return store, nil
}

// AddStore adds an existing Store instance to the StoreManager for testing purposes
func (s *Stores) AddExistingStore(store modules.Store) error {
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

// CloseAllStores closes all the stores in the StoreManager
func (s *Stores) CloseAllStores() error {
	s.m.Lock()
	defer s.m.Unlock()
	for _, store := range s.stores {
		if err := store.Stop(); err != nil {
			return err
		}
	}
	return nil
}
