package store

import (
	"bytes"
	"crypto/sha256"

	ics23 "github.com/cosmos/ics23/go"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/smt"
)

type position int

const (
	left  position = iota // 0
	right                 // 1
)

var (
	// Custom SMT spec as the store does not hash values
	smtSpec *ics23.ProofSpec = &ics23.ProofSpec{
		LeafSpec: &ics23.LeafOp{
			Hash:         ics23.HashOp_SHA256,
			PrehashKey:   ics23.HashOp_SHA256,
			PrehashValue: ics23.HashOp_NO_HASH,
			Length:       ics23.LengthOp_NO_PREFIX,
			Prefix:       []byte{0},
		},
		InnerSpec: &ics23.InnerSpec{
			ChildOrder:      []int32{0, 1},
			ChildSize:       32,
			MinPrefixLength: 1,
			MaxPrefixLength: 1,
			EmptyChild:      make([]byte, 32),
			Hash:            ics23.HashOp_SHA256,
		},
		MaxDepth:                   256,
		PrehashKeyBeforeComparison: true,
	}
	innerPrefix = []byte{1}

	// defaultValue is the default placeholder value in a SparseMerkleTree
	defaultValue = make([]byte, 32)
)

// VerifyMembership verifies the CommitmentProof provided, checking whether it produces the same
// root as the one given. If it does, the key-value pair is a member of the tree
func VerifyMembership(root *coreTypes.CommitmentRoot, proof *ics23.CommitmentProof, key, value []byte) bool {
	// verify the proof
	return ics23.VerifyMembership(smtSpec, root.Hash, proof, key, value)
}

// VerifyNonMembership verifies the CommitmentProof provided, checking whether it produces the same
// root as the one given. If it does, the key-value pair is not a member of the tree as the proof's
// value is either the default nil value for the SMT or an unrelated value at the path
func VerifyNonMembership(root *coreTypes.CommitmentRoot, proof *ics23.CommitmentProof, key []byte) bool {
	// Verify the proof of the non-membership data doesn't belong to the key
	valid := ics23.VerifyMembership(smtSpec, root.Hash, proof, key, proof.GetExist().GetValue())
	// Verify the key was actually empty
	if bytes.Equal(proof.GetExist().GetValue(), defaultValue) {
		return valid
	}
	// Verify the key was present with unrelated data
	return !valid
}

// createMembershipProof generates a CommitmentProof object verifying the membership of a key-value pair
// in the SMT provided
func createMembershipProof(tree *smt.SMT, key, value []byte) (*ics23.CommitmentProof, error) {
	proof, err := tree.Prove(key)
	if err != nil {
		return nil, coreTypes.ErrCreatingProof(err.Error())
	}
	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: convertSMPToExistenceProof(&proof, key, value),
		},
	}, nil
}

// createNonMembershipProof generates a CommitmentProof object verifying the membership of an unrealted key at the given key in the SMT provided
func createNonMembershipProof(tree *smt.SMT, key []byte) (*ics23.CommitmentProof, error) {
	proof, err := tree.Prove(key)
	if err != nil {
		return nil, coreTypes.ErrCreatingProof(err.Error())
	}

	value := defaultValue
	if proof.NonMembershipLeafData != nil {
		value = proof.NonMembershipLeafData[33:]
	}

	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: convertSMPToExistenceProof(&proof, key, value),
		},
	}, nil
}

// convertSMPToExistenceProof converts a SparseMerkleProof to an ICS23 ExistenceProof used for
// both membership and non-membership proof verification
func convertSMPToExistenceProof(proof *smt.SparseMerkleProof, key, value []byte) *ics23.ExistenceProof {
	path := sha256.Sum256(key)
	steps := make([]*ics23.InnerOp, 0, len(proof.SideNodes))
	for i := 0; i < len(proof.SideNodes); i++ {
		var prefix, suffix []byte
		prefix = append(prefix, innerPrefix...)
		if getPathBit(path[:], len(proof.SideNodes)-1-i) == left {
			suffix = make([]byte, 0, len(proof.SideNodes[i]))
			suffix = append(suffix, proof.SideNodes[i]...)
		} else {
			prefix = append(prefix, proof.SideNodes[i]...)
		}
		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: prefix,
			Suffix: suffix,
		}
		steps = append(steps, op)
	}
	leaf := &ics23.LeafOp{
		Hash:         ics23.HashOp_SHA256,
		PrehashKey:   ics23.HashOp_SHA256,
		PrehashValue: ics23.HashOp_NO_HASH,
		Length:       ics23.LengthOp_NO_PREFIX,
		Prefix:       []byte{0},
	}
	return &ics23.ExistenceProof{
		Key:   key,
		Value: value,
		Leaf:  leaf,
		Path:  steps,
	}
}

// getPathBit determines whether the hash of a node at a certain depth in the tree is the
// left or the right child of its parent
func getPathBit(data []byte, position int) position {
	if int(data[position/8])&(1<<(8-1-uint(position)%8)) > 0 {
		return right
	}
	return left
}
