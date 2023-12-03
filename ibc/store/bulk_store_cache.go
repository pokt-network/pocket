package store

import (
	"path/filepath"
	"sync"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ modules.BulkStoreCacher = &bulkStoreCache{}

	cacheDirs = func(storesDir string) string { return filepath.Join(storesDir, "caches") }
)

type lockableStoreMap struct {
	m      sync.Mutex
	stores map[string]modules.ProvableStore
}

// bulkStoreCache holds an in-memory map of all the provable stores in use
// RESEARCH: Look into  parallelising the caching methods
type bulkStoreCache struct {
	base_modules.IntegrableModule

	cfg        *configs.BulkStoreCacherConfig
	logger     *modules.Logger
	storesDir  string
	privateKey string

	ls *lockableStoreMap
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
	s.ls = &lockableStoreMap{
		m:      sync.Mutex{},
		stores: make(map[string]modules.ProvableStore),
	}
	return s, nil
}

func (s *bulkStoreCache) GetModuleName() string { return modules.BulkStoreCacherModuleName }

// AddStore creates and adds a provableStore to the bulkStoreCache
// if one of the same name does not already exist
func (s *bulkStoreCache) AddStore(name string) error {
	s.ls.m.Lock()
	defer s.ls.m.Unlock()
	if _, ok := s.ls.stores[name]; ok {
		return coreTypes.ErrIBCStoreAlreadyExists(name)
	}
	store := NewProvableStore(s.GetBus(), coreTypes.CommitmentPrefix(name), s.privateKey)
	s.ls.stores[store.name] = store
	return nil
}

// GetStore returns the provableStore with the given name
func (s *bulkStoreCache) GetStore(name string) (modules.ProvableStore, error) {
	s.ls.m.Lock()
	defer s.ls.m.Unlock()
	store, ok := s.ls.stores[name]
	if !ok {
		return nil, coreTypes.ErrIBCStoreDoesNotExist(name)
	}
	return store, nil
}

// RemoveStore removes the provableStore with the given name
func (s *bulkStoreCache) RemoveStore(name string) error {
	s.ls.m.Lock()
	defer s.ls.m.Unlock()
	if _, ok := s.ls.stores[name]; !ok {
		return coreTypes.ErrIBCStoreDoesNotExist(name)
	}
	delete(s.ls.stores, name)
	return nil
}

// GetAllStores returns the map of stores to their store names
func (s *bulkStoreCache) GetAllStores() map[string]modules.ProvableStore {
	return s.ls.stores
}

// FlushdCachesToStore caches all the entries for all stores in the bulkStoreCache
func (s *bulkStoreCache) FlushCachesToStore() error {
	s.ls.m.Lock()
	defer s.ls.m.Unlock()
	s.logger.Info().Msg("🚽 Flushing All Cache Entries to Disk 🚽")
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.ls.stores {
		if err := store.FlushCache(disk); err != nil {
			s.logger.Error().Err(err).Str("store", string(store.GetCommitmentPrefix())).Msg("🚨 Error Flushing Cache 🚨")
			return err
		}
	}
	return disk.Stop()
}

// PruneCaches prunes the caches for all stores in the bulkStoreCache at the given height
func (s *bulkStoreCache) PruneCaches(height uint64) error {
	s.ls.m.Lock()
	defer s.ls.m.Unlock()
	s.logger.Info().Uint64("height", height).Msg("✂️  Pruning Cache Entries at Height ✂️ ")
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.ls.stores {
		if err := store.PruneCache(disk, height); err != nil {
			s.logger.Error().Err(err).Str("store", string(store.GetCommitmentPrefix())).Msg("🚨 Error Pruning Cache 🚨")
			return err
		}
	}
	return disk.Stop()
}

// RestoreCaches restores the caches from disk for all stores in the bulkStoreCache
func (s *bulkStoreCache) RestoreCaches(height uint64) error {
	s.ls.m.Lock()
	defer s.ls.m.Unlock()
	s.logger.Info().Msg("📥 Restoring Cache Entries from Disk 📥")
	disk, err := newKVStore(s.storesDir)
	if err != nil {
		return err
	}
	for _, store := range s.ls.stores {
		if err := store.RestoreCache(disk, height); err != nil {
			s.logger.Error().Err(err).Str("store", string(store.GetCommitmentPrefix())).Msg("🚨 Error Restoring Cache 🚨")
			return err
		}
	}
	return disk.Stop()
}

func newKVStore(dir string) (kvstore.KVStore, error) {
	if dir == ":memory:" {
		return kvstore.NewMemKVStore(), nil
	}
	return kvstore.NewKVStore(cacheDirs(dir))
}
