package store

import (
	"bytes"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/ibc/host"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.ProvableStore = &provableStore{}

// provableStore implements the ProvableStore interface and wraps an SMT
// it operates in memory and thus cannot make any changes to the underlying
// database. All changes must be propagated through the `IbcMessage` type
// and added to the mempool for inclusion in the next block
type provableStore struct {
	prefix coreTypes.CommitmentPrefix
	tree   *smt.SMT
}

// GetCommitmentPrefix returns the commitment prefix of the provable store
func (p *provableStore) GetCommitmentPrefix() coreTypes.CommitmentPrefix {
	return p.prefix
}

// Get returns the value in the tree of the key prefixed with the CommitmentPrefix
func (p *provableStore) Get(key []byte) ([]byte, error) {
	prefixed := applyPrefix(p.prefix, key)
	return p.tree.Get(prefixed)
}

// Set sets the value in the tree of the key prefixed with the CommitmentPrefix
func (p *provableStore) Set(key, value []byte) error {
	prefixed := applyPrefix(p.prefix, key)
	return p.tree.Update(prefixed, value)
}

// Delete deletes the value in the tree of the key prefixed with the CommitmentPrefix
func (p *provableStore) Delete(key []byte) error {
	prefixed := applyPrefix(p.prefix, key)
	return p.tree.Delete(prefixed)
}

// CreateMembershipProof creates a membership proof for the key-value pair with the key
// prefixed with the CommitmentPrefix
func (p *provableStore) CreateMembershipProof(key, value []byte) (*ics23.CommitmentProof, error) {
	prefixed := applyPrefix(p.prefix, key)
	return createMembershipProof(p.tree, prefixed, value)
}

// CreateNonMembershipProof creates a non-membership proof for the key prefixed with the CommitmentPrefix
func (p *provableStore) CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	prefixed := applyPrefix(p.prefix, key)
	return createNonMembershipProof(p.tree, prefixed)
}

// Root returns the root of the entire tree
// NOTE: Root does not work on a per-prefix basis but returns the root of the entire tree
func (p *provableStore) Root() []byte {
	return p.tree.Root()
}

// applyPrefix will apply the CommitmentPrefix to the key provided if not already applied
func applyPrefix(prefix coreTypes.CommitmentPrefix, key []byte) coreTypes.CommitmentPath {
	slashed := make([]byte, 0, len(key)+1)
	slashed = append(slashed, key...)
	slashed = append(slashed, []byte("/")...)
	if bytes.Equal(prefix[:len(slashed)], slashed) {
		return key
	}
	return host.ApplyPrefix(prefix, string(key))
}
