package stores

import (
	"crypto/sha256"

	"github.com/pokt-network/pocket/ibc/host"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

// NewTestStore creates a new store for testing purposes using an in memory KVStore
func NewTestStore(storeKey string, provable bool) (modules.Store, error) {
	db := kvstore.NewMemKVStore()
	if !provable {
		return &PrivateStore{db, storeKey}, nil
	}
	// Create a new SMT with no value hasher to store the unhashed value bytes in the tree
	tree := smt.NewSparseMerkleTree(db, sha256.New(), noValueHasher)
	return &ProvableStore{db, tree, storeKey}, nil
}

// NewStore creates a new store using a persistent KVStore at the path provided.  Stores can either
// be provable or not provable, a non provable store will simply be a KVStore instance, a provable
// store will be a KVStore instance with an SMT on top of it for proof verification
func NewStore(storeKey, storePath string, provable bool) (modules.Store, error) {
	db, err := kvstore.NewKVStore(storePath)
	if err != nil {
		return nil, coreTypes.ErrStoreCreation(err)
	}
	if !provable {
		return &PrivateStore{db, storeKey}, nil
	}
	// Create a new SMT with no value hasher to store the unhashed value bytes in the tree
	tree := smt.NewSparseMerkleTree(db, sha256.New(), noValueHasher)
	return &ProvableStore{db, tree, storeKey}, nil
}

// PopulateStore populates a store with the data provided by iterating through the map,
// applying the store prefix to the paths and setting the value in the store to this key
func PopulateStore(store modules.Store, data map[string][]byte) error {
	for path, value := range data {
		prefix := &coreTypes.CommitmentPrefix{Prefix: []byte(store.GetStoreKey())}
		key := host.ApplyPrefix(prefix, path).GetPath()
		if err := store.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}
