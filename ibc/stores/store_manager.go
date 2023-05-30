package stores

import (
	"sync"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
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

// GetStore retrieves a Store instance from the StoreManager
func (s *Stores) GetStore(storeKey string) (modules.Store, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.stores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	return store, nil
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
