package store

import (
	"crypto/sha256"

	ics23 "github.com/cosmos/ics23/go"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/smt"
)

// position refers to whether the node is either the left or right child of its parent
// for the binary SMT
// Ref: https://github.com/pokt-network/smt/blob/main/types.go
const (
	left     int = iota // 0
	right               // 1
	hashSize = 32
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
			ChildSize:       hashSize,
			MinPrefixLength: 1,
			MaxPrefixLength: 1,
			EmptyChild:      make([]byte, hashSize),
			Hash:            ics23.HashOp_SHA256,
		},
		MaxDepth:                   256,
		PrehashKeyBeforeComparison: true,
	}
	innerPrefix = []byte{1}

	// defaultValue is the default placeholder value in a SparseMerkleTree
	defaultValue = make([]byte, hashSize)
)

// VerifyMembership verifies the CommitmentProof provided, checking whether it produces the same
// root as the one given. If it does, the key-value pair is a member of the tree
func VerifyMembership(root ics23.CommitmentRoot, proof *ics23.CommitmentProof, key, value []byte) bool {
	// verify the proof
	return ics23.VerifyMembership(smtSpec, root, proof, key, value)
}

// VerifyNonMembership verifies the CommitmentProof provided, checking whether it produces the same
// root as the one given. If it does, the key-value pair is not a member of the tree as the proof's
// value is either the default nil value for the SMT or an unrelated value at the path
func VerifyNonMembership(root ics23.CommitmentRoot, proof *ics23.CommitmentProof, key []byte) bool {
	// verify the proof
	return ics23.VerifyNonMembership(smtSpec, root, proof, key)
}

// createMembershipProof generates a CommitmentProof object verifying the membership of a key-value pair
// in the SMT provided
func createMembershipProof(tree *smt.SMT, key, value []byte) (*ics23.CommitmentProof, error) {
	proof, err := tree.Prove(key)
	if err != nil || proof == nil {
		return nil, coreTypes.ErrCreatingProof(err)
	}
	return convertSMPToExistenceProof(proof, key, value), nil
}

// createNonMembershipProof generates a CommitmentProof object verifying the membership of an unrealted key at the given key in the SMT provided
func createNonMembershipProof(tree *smt.SMT, key []byte) (*ics23.CommitmentProof, error) {
	proof, err := tree.Prove(key)
	if err != nil || proof == nil {
		return nil, coreTypes.ErrCreatingProof(err)
	}
	return convertSMPToExclusionProof(proof, key), nil
}

// convertSMPToExistenceProof converts a SparseMerkleProof to an ics23
// ExistenceProof to verify membership of an element
func convertSMPToExistenceProof(proof *smt.SparseMerkleProof, key, value []byte) *ics23.CommitmentProof {
	path := sha256.Sum256(key)
	steps := convertSideNodesToSteps(proof.SideNodes, path[:])
	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: &ics23.ExistenceProof{
				Key:   key,
				Value: value,
				Leaf:  smtSpec.LeafSpec,
				Path:  steps,
			},
		},
	}
}

// convertSMPToExclusionProof converts a SparseMerkleProof to an ics23
// ExclusionProof to verify non-membership of an element
func convertSMPToExclusionProof(proof *smt.SparseMerkleProof, key []byte) *ics23.CommitmentProof {
	path := sha256.Sum256(key)
	steps := convertSideNodesToSteps(proof.SideNodes, path[:])
	leaf := &ics23.LeafOp{
		Hash: ics23.HashOp_SHA256,
		// Do not re-hash already hashed fields from NonMembershipLeafData
		PrehashKey:   ics23.HashOp_NO_HASH,
		PrehashValue: ics23.HashOp_NO_HASH,
		Length:       ics23.LengthOp_NO_PREFIX,
		Prefix:       []byte{0},
	}
	actualPath := path[:]
	actualValue := defaultValue
	if proof.NonMembershipLeafData != nil {
		actualPath = proof.NonMembershipLeafData[1 : 1+hashSize] // len(prefix): len(prefix) + hashSize
		actualValue = proof.NonMembershipLeafData[1+hashSize:]
	}
	return &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exclusion{
			Exclusion: &ics23.ExclusionProof{
				Key:             key,
				ActualPath:      actualPath,
				ActualValueHash: actualValue,
				Leaf:            leaf,
				Path:            steps,
			},
		},
	}
}

// convertSideNodesToSteps converts the SideNodes field in the SparseMerkleProof
// into a list of InnerOps for the ics23 CommitmentProof
func convertSideNodesToSteps(sideNodes [][]byte, path []byte) []*ics23.InnerOp {
	steps := make([]*ics23.InnerOp, 0, len(sideNodes))
	for i := 0; i < len(sideNodes); i++ {
		var prefix, suffix []byte
		prefix = append(prefix, innerPrefix...)
		if isLeft(path, len(sideNodes)-1-i) {
			// path is on the left so sidenode must be on the right
			suffix = make([]byte, 0, len(sideNodes[i]))
			suffix = append(suffix, sideNodes[i]...)
		} else {
			// path is on the right so sidenode must be on the left
			prefix = append(prefix, sideNodes[i]...)
		}
		op := &ics23.InnerOp{
			Hash:   ics23.HashOp_SHA256,
			Prefix: prefix,
			Suffix: suffix,
		}
		steps = append(steps, op)
	}
	return steps
}

// isLeft returns true is the i-th bit of path is a left child in the SMT
func isLeft(path []byte, i int) bool {
	return smt.GetPathBit(path, i) == left
}
