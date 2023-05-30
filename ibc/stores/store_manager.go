package stores

import (
	"crypto/sha256"
	"sync"

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

func NewStoreManager() modules.StoreManager {
	return &Stores{
		stores: make(map[string]modules.Store),
	}
}

func NewStore(storeKey, storePath string, provable bool) (modules.Store, error) {
	db, err := kvstore.NewKVStore(storePath)
	if err != nil {
		return nil, coreTypes.ErrStoreCreation(err)
	}
	if !provable {
		return &PrivateStore{db, storeKey}, nil
	}
	// Create a new SMT with no value hasher to store the unhashed value bytes in the tree
	smt := smt.NewSparseMerkleTree(db, sha256.New(), smt.WithValueHasher(nil))
	return &ProvableStore{db, smt, storeKey}, nil
}

func (s *Stores) GetStore(storeKey string) (modules.Store, error) {
	s.m.Lock()
	defer s.m.Unlock()
	store, ok := s.stores[storeKey]
	if !ok {
		return nil, coreTypes.ErrStoreNotFound(storeKey)
	}
	return store, nil
}

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

func (s *Stores) RemoveStore(storeKey string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.stores[storeKey]; !ok {
		return coreTypes.ErrStoreNotFound(storeKey)
	}
	delete(s.stores, storeKey)
	return nil
}
