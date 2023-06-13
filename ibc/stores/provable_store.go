package stores

import (
	"crypto/sha256"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var (
	_             modules.ProvableStore = &ProvableStore{}
	noValueHasher                       = smt.WithValueHasher(nil)
)

type CommitmentRoot []byte

// ProvableStore needs to produce CommitmentProof objects verifying membership
// and non-membership of keys in the store, as such the ProvableStore utilises
// a Sparse Merkle Tree (SMT) to store the data and generate such proofs
type ProvableStore struct {
	nodeStore kvstore.KVStore
	tree      *smt.SMT
	storeKey  string
	provable  bool
}

// newProvableStoreFromKV generates a new provable store from the nodeStore provided
func newProvableStoreFromKV(nodeStore kvstore.KVStore, storeKey string) *ProvableStore {
	tree := smt.NewSparseMerkleTree(nodeStore, sha256.New(), noValueHasher)
	return &ProvableStore{
		nodeStore: nodeStore,
		tree:      tree,
		storeKey:  storeKey,
		provable:  true,
	}
}

// NewTestProvableStore generates a new provable store for testing purposes
func NewTestProvableStore(storeKey string, nodeStore kvstore.KVStore) modules.ProvableStore {
	if nodeStore == nil {
		ns := kvstore.NewMemKVStore()
		return newProvableStoreFromKV(ns, storeKey)
	}
	return newProvableStoreFromKV(nodeStore, storeKey)
}

func (prov *ProvableStore) GetStoreKey() string {
	return prov.storeKey
}

func (prov *ProvableStore) IsProvable() bool {
	return prov.provable
}

func (prov *ProvableStore) Get(key []byte) ([]byte, error) {
	return prov.tree.Get(key)
}

// Set atomically updates the tree reverting to the previous state if any error occurs
// during the update or underlying database commit
func (prov *ProvableStore) Set(key, value []byte) error {
	pre := smt.ImportSparseMerkleTree(prov.nodeStore, sha256.New(), prov.tree.Root(), noValueHasher)
	if err := prov.tree.Update(key, value); err != nil {
		prov.tree = pre
		return coreTypes.ErrIBCStoreUpdate(err)
	}
	if err := prov.tree.Commit(); err != nil {
		prov.tree = pre
		return coreTypes.ErrIBCStoreUpdate(err)
	}
	return nil
}

// Delete atomically deletes from the tree reverting to the previous state if any error occurs
// during the update or underlying database commit
func (prov *ProvableStore) Delete(key []byte) error {
	pre := smt.ImportSparseMerkleTree(prov.nodeStore, sha256.New(), prov.tree.Root(), noValueHasher)
	if err := prov.tree.Delete(key); err != nil {
		prov.tree = pre
		return coreTypes.ErrIBCStoreUpdate(err)
	}
	if err := prov.tree.Commit(); err != nil {
		prov.tree = pre
		return coreTypes.ErrIBCStoreUpdate(err)
	}
	return nil
}

// Stop closes the undelying database behind the SMT
func (prov *ProvableStore) Stop() error {
	return prov.nodeStore.Stop()
}

// Root returns the root of the SMT as a CommitmentRoot object
func (prov *ProvableStore) Root() *coreTypes.CommitmentRoot {
	return &coreTypes.CommitmentRoot{Hash: prov.tree.Root()}
}

// CreateMembershipProof generates a CommitmentProof object verifying the membership of a key-value pair
func (prov *ProvableStore) CreateMembershipProof(key, value []byte) (*ics23.CommitmentProof, error) {
	return createMembershipProof(prov.tree, key, value)
}

// CreateNonMembershipProof generates a CommitmentProof object verifying the non-membership of a key
func (prov *ProvableStore) CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	return createNonMembershipProof(prov.tree, key)
}
