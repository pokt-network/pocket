package stores

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.ProvableStore = (*ProvableStore)(nil)

// ProvableStore needs to produce CommitmentProof objects verifying membership
// and non-membership of keys in the store, as such the ProvableStore utilises
// a Sparse Merkle Tree (SMT) to store the data and generate such proofs
type ProvableStore struct {
	nodeStore kvstore.KVStore
	tree      *smt.SMT
	storeKey  string
}

func (prov *ProvableStore) GetStoreKey() string {
	return prov.storeKey
}

func (prov *ProvableStore) Get(key []byte) ([]byte, error) {
	return prov.tree.Get(key)
}

func (prov *ProvableStore) Set(key, value []byte) error {
	return prov.tree.Update(key, value)
}

func (prov *ProvableStore) Delete(key []byte) error {
	return prov.tree.Delete(key)
}

func (prov *ProvableStore) Root() []byte {
	return prov.tree.Root()
}

func (prov *ProvableStore) CreateMembershipProof(key, value []byte) (*coreTypes.CommitmentProof, error) {
	return generateProof(prov.tree, key, value)
}

func (prov *ProvableStore) CreateNonMembershipProof(key []byte) (*coreTypes.CommitmentProof, error) {
	return generateProof(prov.tree, key, nil)
}

func generateProof(tree *smt.SMT, key, value []byte) (*coreTypes.CommitmentProof, error) {
	smtProof, err := tree.Prove(key)
	if err != nil {
		return nil, err
	}
	proof := &coreTypes.CommitmentProof{
		Key:                   key,
		Value:                 value,
		SideNodes:             smtProof.SideNodes,
		NonMembershipLeafData: smtProof.NonMembershipLeafData,
		SiblingData:           smtProof.SiblingData,
	}
	return proof, nil
}
