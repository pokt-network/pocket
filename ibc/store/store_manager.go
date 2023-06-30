package store

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var (
	_         modules.IBCStoreManager = &storeManager{}
	cacheDirs                         = func(storesDir string) string { return fmt.Sprintf("%s/caches", storesDir) }
)

// storeManager holds an in-memory map of all the provable stores in use
type storeManager struct {
	m         sync.Mutex
	bus       modules.Bus
	storesDir string
	stores    map[string]*provableStore
}

// NewStoreManager returns a new storeManager instance
func NewStoreManager(bus modules.Bus, storesDir string) *storeManager {
	return &storeManager{
		m:         sync.Mutex{},
		bus:       bus,
		storesDir: storesDir,
		stores:    make(map[string]*provableStore, 0),
	}
}

// AddStore creates and adds a provableStore to the storeManager
// if one of the same name does not already exist
func (s *storeManager) AddStore(name string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[name]; ok {
		return coreTypes.ErrIBCStoreAlreadyExists(name)
	}
	store := newProvableStore(s.bus, coreTypes.CommitmentPrefix(name))
	s.stores[store.name] = store
	return nil
}

// GetStore returns the provableStore with the given name
func (s *storeManager) GetStore(name string) (modules.ProvableStore, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.stores[name]
	if !ok {
		return nil, coreTypes.ErrIBCStoreDoesNotExist(name)
	}
	return store, nil
}

// RemoveStore removes the provableStore with the given name
func (s *storeManager) RemoveStore(name string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[name]; !ok {
		return coreTypes.ErrIBCStoreDoesNotExist(name)
	}
	delete(s.stores, name)
	return nil
}

// GetAllStores returns the map of stores to their store names
func (s *storeManager) GetAllStores() map[string]modules.ProvableStore {
	s.m.Lock()
	defer s.m.Unlock()
	stores := make(map[string]modules.ProvableStore, len(s.stores))
	for name, store := range s.stores {
		stores[name] = store
	}
	return stores
}

// CacheAllEntries caches all the entries for all stores in the storeManager
func (s *storeManager) CacheAllEntries() error {
	s.m.Lock()
	defer s.m.Unlock()
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.stores {
		if err := store.CacheEntries(disk); err != nil {
			return err
		}
	}
	return disk.Stop()
}

// PruneCaches prunes the caches for all stores in the storeManager at the given height
func (s *storeManager) PruneCaches(height uint64) error {
	s.m.Lock()
	defer s.m.Unlock()
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.stores {
		if err := store.PruneCache(disk, height); err != nil {
			return err
		}
	}
	return disk.Stop()
}

// RestoreCaches restores the caches from disk for all stores in the storeManager
func (s *storeManager) RestoreCaches() error {
	s.m.Lock()
	defer s.m.Unlock()
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.stores {
		if err := store.RestoreCache(disk); err != nil {
			return err
		}
	}
	return disk.Stop()
}

func newKVStore(dir string) (kvstore.KVStore, error) {
	if dir == ":memory:" {
		return kvstore.NewMemKVStore(), nil
	}
	store, err := kvstore.NewKVStore(cacheDirs(dir))
	if err != nil {
		return nil, err
	}
	return store, nil
}
