package modules

import (
	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

//go:generate mockgen -destination=./mocks/ibc_store_module_mock.go github.com/pokt-network/pocket/shared/modules BulkStoreCacher,ProvableStore

const BulkStoreCacherModuleName = "bulk_store_cache"

type BulkStoreCacherOption func(BulkStoreCacher)

type bulkStoreCacheFactory = FactoryWithConfigAndOptions[BulkStoreCacher, *configs.BulkStoreCacherConfig, BulkStoreCacherOption]

// BulkStoreCacher is a submodule that interacts with the different stores it manages and handles the
// flushing, pruning and restoration of their caches in bulk
type BulkStoreCacher interface {
	Submodule
	bulkStoreCacheFactory

	AddStore(name string) error
	GetStore(name string) (ProvableStore, error)
	RemoveStore(name string) error
	GetAllStores() map[string]ProvableStore
	FlushCachesToStore() error
	PruneCaches(height uint64) error
	RestoreCaches(height uint64) error
}

// ProvableStore allows for the Get/Set/Delete operations as well as the generation of proofs for
// any element in the store (or not in the store). Its provable nature allows for the retrieval of
// the root hash of the underlying tree structure that backs the ProvableStore instance.
// The ProvableStore also maintains a cache of any changes it makes to the underlying store, which
// can be flushed to a separate database, pruned and restored when necessary.
type ProvableStore interface {
	Get(key []byte) ([]byte, error)
	GetAndProve(key []byte) ([]byte, *ics23.CommitmentProof, error)
	CreateMembershipProof(key, value []byte) (*ics23.CommitmentProof, error)
	CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error)
	Set(key, value []byte) error
	Delete(key []byte) error
	GetCommitmentPrefix() coreTypes.CommitmentPrefix
	Root() ics23.CommitmentRoot
	FlushCache(kvstore.KVStore) error
	PruneCache(store kvstore.KVStore, height uint64) error
	RestoreCache(store kvstore.KVStore, height uint64) error
}
