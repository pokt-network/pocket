package stores

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.StoreManager = (*Stores)(nil)

type Stores struct {
	stores map[string]modules.Store
}

func NewStoreManager() modules.StoreManager {
	return &Stores{
		stores: make(map[string]modules.Store),
	}
}

func (s *Stores) GetStore(storeKey string) (modules.Store, error) {
	store, ok := s.stores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	return store, nil
}

func (s *Stores) AddStore(store modules.Store, storeKey string) error {
	if _, ok := s.stores[storeKey]; ok {
		return coreTypes.ErrStoreAlreadyExists(storeKey)
	}
	s.stores[storeKey] = store
	return nil
}

func (s *Stores) RemoveStore(storeKey string) error {
	if _, ok := s.stores[storeKey]; !ok {
		return coreTypes.ErrStoreNotFound(storeKey)
	}
	delete(s.stores, storeKey)
	return nil
}
