package modules

import (
	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"google.golang.org/protobuf/types/known/anypb"
)

//go:generate mockgen -destination=./mocks/ibc_module_mock.go github.com/pokt-network/pocket/shared/modules IBCModule,IBCStoreManager,ProvableStore

const IBCModuleName = "ibc"

type IBCModule interface {
	Module

	// HandleMessage handles the given IBC message
	HandleMessage(*anypb.Any) error
}

// IBCStoreManager manages the different ProvableStore instances created by the IBC host
type IBCStoreManager interface {
	AddStore(name string) error
	GetStore(name string) (ProvableStore, error)
	RemoveStore(name string) error
	GetAllStores() map[string]ProvableStore
	CacheAllEntries() error
	PruneCaches(height uint64) error
	RestoreCaches() error
}

// ProvableStore interacts with Persistence and the IBC state tree in order for the IBC host to
// be able to interact with the IBC store locally and propagate any changes throuhout the network
type ProvableStore interface {
	Get(key []byte) ([]byte, error)
	GetAndProve(key []byte, membership bool) ([]byte, *ics23.CommitmentProof, error)
	CreateMembershipProof(key, value []byte) (*ics23.CommitmentProof, error)
	CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error)
	Set(key, value []byte) error
	Delete(key []byte) error
	GetCommitmentPrefix() coreTypes.CommitmentPrefix
	Root() ics23.CommitmentRoot
	FlushEntries(kvstore.KVStore) error
	PruneCache(store kvstore.KVStore, height uint64) error
	RestoreCache(kvstore.KVStore) error
}
