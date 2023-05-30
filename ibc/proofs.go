package ibc

import (
	"bytes"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/smt"
)

// defaultValue is the default placeholder value in a SparseMerkleTree
var defaultValue []byte = nil

// validateMembershipProof validates the proof for membership by doing basic validation
// For a MembershipProof, the key-value pair should match the proof's embedded key-value
// pair and the NonMembershipLeafData should be nil.
func validateMembershipProof(proof *coreTypes.CommitmentProof, key, value []byte) error {
	if !bytes.Equal(proof.GetKey(), key) {
		return coreTypes.ErrInvalidProof("provided key does not match proof key")
	}
	if !bytes.Equal(proof.GetValue(), value) {
		return coreTypes.ErrInvalidProof("provided value does not match proof value")
	}
	if proof.NonMembershipLeafData != nil {
		return coreTypes.ErrInvalidProof("non membership leaf data must be nil")
	}
	return nil
}

// validateNonMembershipProof validates the proof for non-membership by doing basic validation
// For a NonMembershipProof, the value should be nil and the NonMembershipLeafData should not be nil
func validateNonMembershipProof(proof *coreTypes.CommitmentProof, key []byte) error {
	if !bytes.Equal(proof.GetKey(), key) {
		return coreTypes.ErrInvalidProof("provided key does not match proof key")
	}
	if !bytes.Equal(proof.GetValue(), defaultValue) {
		return coreTypes.ErrInvalidProof("default value does not match proof value")
	}
	if proof.NonMembershipLeafData == nil {
		return coreTypes.ErrInvalidProof("non membership leaf data must not be nil")
	}
	return nil
}

// verifyCommitmentProof is a wrapper that converts the protobuf CommitmentProof to a SparseMerkleProof
// in order to verify the proof using the SMT library.
func verifyCommitmentProof(spec *smt.TreeSpec, proof *coreTypes.CommitmentProof, root, key, value []byte) (bool, error) {
	smtProof := smt.SparseMerkleProof{
		SideNodes:             proof.SideNodes,
		NonMembershipLeafData: proof.NonMembershipLeafData,
		SiblingData:           proof.SiblingData,
	}
	return smt.VerifyProof(smtProof, root, key, value, spec), nil
}

// VerifyMembership verifies whether a given key-value pair is contained in the tree according
// to the proof provided. It does so by converting the CommitmentProof to a SparseMerkleProof
// and then rebuilds the root hash of the tree. If the root hash matches the one provided, the
// key-value pair is contained in the tree.
func VerifyMembership(spec *smt.TreeSpec, proof *coreTypes.CommitmentProof, root, key, value []byte) (bool, error) {
	if err := validateMembershipProof(proof, key, value); err != nil {
		return false, err
	}
	return verifyCommitmentProof(spec, proof, root, key, value)
}

// VerifyNonMembership verifies whether a given key is not contained in the tree according
// to the proof provided. It does so by converting the CommitmentProof to a SparseMerkleProof
// and then rebuilds the root hash of the tree using the unrelated key-value pair in the
// position of the provided key in the tree. If the root hash matches the one provided, the
// key is not contained in the tree.
func VerifyNonMembership(spec *smt.TreeSpec, proof *coreTypes.CommitmentProof, root, key []byte) (bool, error) {
	if err := validateNonMembershipProof(proof, key); err != nil {
		return false, err
	}
	return verifyCommitmentProof(spec, proof, root, key, defaultValue)
}
