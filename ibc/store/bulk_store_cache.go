package store

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ modules.BulkStoreCacher = &bulkStoreCache{}

	cacheDirs = func(storesDir string) string { return fmt.Sprintf("%s/caches", storesDir) }
)

// bulkStoreCache holds an in-memory map of all the provable stores in use
type bulkStoreCache struct {
	base_modules.IntegrableModule

	m sync.Mutex

	cfg        *configs.BulkStoreCacherConfig
	logger     *modules.Logger
	storesDir  string
	privateKey string

	stores map[string]*provableStore
}

func Create(bus modules.Bus, config *configs.BulkStoreCacherConfig, options ...modules.BulkStoreCacherOption) (modules.BulkStoreCacher, error) {
	return new(bulkStoreCache).Create(bus, config, options...)
}

// WithLogger assigns a logger for the bulk store cache
func WithLogger(logger *modules.Logger) modules.BulkStoreCacherOption {
	return func(m modules.BulkStoreCacher) {
		if mod, ok := m.(*bulkStoreCache); ok {
			mod.logger = logger
		}
	}
}

// WithPrivateKey assigns the private key to the BulkStoreCacher
func WithPrivateKey(privateKeyHex string) modules.BulkStoreCacherOption {
	return func(m modules.BulkStoreCacher) {
		if mod, ok := m.(*bulkStoreCache); ok {
			mod.privateKey = privateKeyHex
		}
	}
}

// WithStoresDir assigns the stores directory to the BulkStoreCacher
func WithStoresDir(storesDir string) modules.BulkStoreCacherOption {
	return func(m modules.BulkStoreCacher) {
		if mod, ok := m.(*bulkStoreCache); ok {
			mod.storesDir = storesDir
		}
	}
}

func (*bulkStoreCache) Create(bus modules.Bus, config *configs.BulkStoreCacherConfig, options ...modules.BulkStoreCacherOption) (modules.BulkStoreCacher, error) {
	s := &bulkStoreCache{
		cfg: config,
	}
	for _, option := range options {
		option(s)
	}
	s.logger.Info().Msg("💾 Creating Bulk Store Cacher 💾")
	bus.RegisterModule(s)
	s.m = sync.Mutex{}
	s.stores = make(map[string]*provableStore)
	return s, nil
}

func (s *bulkStoreCache) GetModuleName() string { return modules.BulkStoreCacherModuleName }

// AddStore creates and adds a provableStore to the bulkStoreCache
// if one of the same name does not already exist
func (s *bulkStoreCache) AddStore(name string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[name]; ok {
		return coreTypes.ErrIBCStoreAlreadyExists(name)
	}
	store := newProvableStore(s.GetBus(), coreTypes.CommitmentPrefix(name), s.privateKey)
	s.stores[store.name] = store
	return nil
}

// GetStore returns the provableStore with the given name
func (s *bulkStoreCache) GetStore(name string) (modules.ProvableStore, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.stores[name]
	if !ok {
		return nil, coreTypes.ErrIBCStoreDoesNotExist(name)
	}
	return store, nil
}

// RemoveStore removes the provableStore with the given name
func (s *bulkStoreCache) RemoveStore(name string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[name]; !ok {
		return coreTypes.ErrIBCStoreDoesNotExist(name)
	}
	delete(s.stores, name)
	return nil
}

// GetAllStores returns the map of stores to their store names
func (s *bulkStoreCache) GetAllStores() map[string]modules.ProvableStore {
	s.m.Lock()
	defer s.m.Unlock()
	stores := make(map[string]modules.ProvableStore, len(s.stores))
	for name, store := range s.stores {
		stores[name] = store
	}
	return stores
}

// FlushAllEntries caches all the entries for all stores in the bulkStoreCache
func (s *bulkStoreCache) FlushAllEntries() error {
	s.m.Lock()
	defer s.m.Unlock()
	s.logger.Info().Msg("🚽 Flushing All Cache Entries to Disk 🚽")
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.stores {
		if err := store.FlushEntries(disk); err != nil {
			return err
		}
	}
	return disk.Stop()
}

// PruneCaches prunes the caches for all stores in the bulkStoreCache at the given height
func (s *bulkStoreCache) PruneCaches(height uint64) error {
	s.m.Lock()
	defer s.m.Unlock()
	s.logger.Info().Uint64("height", height).Msg("✂️  Pruning Cache Entries at Height ✂️ ")
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

// RestoreCaches restores the caches from disk for all stores in the bulkStoreCache
func (s *bulkStoreCache) RestoreCaches() error {
	s.m.Lock()
	defer s.m.Unlock()
	s.logger.Info().Msg("📥 Restoring Cache Entries from Disk 📥")
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
