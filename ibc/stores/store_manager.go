package stores

//go:generate mockgen -package=mock_types -destination=../types/mocks/store_manager_mock.go github.com/pokt-network/pocket/ibc/stores.privateStores

import (
	"sync"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.StoreManager = (*Stores)(nil)

type Stores struct {
	m              sync.Mutex
	privateStores  map[string]modules.PrivateStore
	provableStores map[string]modules.ProvableStore
}

// NewStoreManager creates a new store manager instance with an empty map of stores
func NewStoreManager() modules.StoreManager {
	return &Stores{
		privateStores:  make(map[string]modules.PrivateStore),
		provableStores: make(map[string]modules.ProvableStore),
	}
}

// GetPrivateStore retrieves a PrivateStore instance from the StoreManager
func (s *Stores) GetPrivateStore(storeKey string) (modules.PrivateStore, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.privateStores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	return store, nil
}

// AddPrivateStore adds a PrivateStore instance to the StoreManager
func (s *Stores) AddPrivateStore(store modules.PrivateStore) error {
	s.m.Lock()
	defer s.m.Unlock()
	storeKey := store.GetStoreKey()
	if _, ok := s.privateStores[storeKey]; ok {
		return coreTypes.ErrStoreAlreadyExists(storeKey)
	}
	s.privateStores[storeKey] = store
	return nil
}

// GetProvableStore retrieves a ProvableStore instance from the StoreManager
func (s *Stores) GetProvableStore(storeKey string) (modules.ProvableStore, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.provableStores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	return store, nil
}

// AddProvableStore adds a ProvableStore instance to the StoreManager
func (s *Stores) AddProvableStore(store modules.ProvableStore) error {
	s.m.Lock()
	defer s.m.Unlock()
	storeKey := store.GetStoreKey()
	if _, ok := s.provableStores[storeKey]; ok {
		return coreTypes.ErrStoreAlreadyExists(storeKey)
	}
	s.provableStores[storeKey] = store
	return nil
}

// RemoveStore removes a Store instance from the StoreManager
func (s *Stores) RemoveStore(storeKey string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.privateStores[storeKey]; ok {
		delete(s.privateStores, storeKey)
		return nil
	}
	if _, ok := s.provableStores[storeKey]; ok {
		delete(s.provableStores, storeKey)
		return nil
	}
	return coreTypes.ErrStoreNotFound(storeKey)
}
