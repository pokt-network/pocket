package store

import (
	"crypto/sha256"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

// Proof generation cannot fail but verification can
func TestICS23Proofs_GenerateCommitmentProofs(t *testing.T) {
	nodeStore := kvstore.NewMemKVStore()
	tree := smt.NewSparseMerkleTree(nodeStore, sha256.New(), smt.WithValueHasher(nil))
	require.NotNil(t, tree)

	// Set a value in the store
	err := tree.Update([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	err = tree.Update([]byte("bar"), []byte("foo"))
	require.NoError(t, err)
	err = tree.Update([]byte("testKey"), []byte("testValue"))
	require.NoError(t, err)
	err = tree.Update([]byte("testKey2"), []byte("testValue2"))
	require.NoError(t, err)

	testCases := []struct {
		key           []byte
		value         []byte
		nonmembership bool
		fails         bool
		expected      error
	}{
		{
			// Successfully generates a membership proof for a key stored
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			fails:         false,
			expected:      nil,
		},
		{
			// Successfully generates a non-membership proof for a key not stored
			key:           []byte("baz"),
			value:         []byte("testValue2"), // unrelated leaf data
			nonmembership: true,
			fails:         false,
			expected:      nil,
		},
		{
			// Successfully generates a non-membership proof for an unset nil key
			key:           nil,
			value:         []byte("foo"), // unrelated leaf data
			nonmembership: true,
			fails:         false,
			expected:      nil,
		},
	}

	for _, tc := range testCases {
		var proof *ics23.CommitmentProof
		if tc.nonmembership {
			proof, err = createNonMembershipProof(tree, tc.key)
		} else {
			proof, err = createMembershipProof(tree, tc.key, tc.value)
		}
		if tc.fails {
			require.EqualError(t, err, tc.expected.Error())
			require.Nil(t, proof)
			return
		}
		require.NoError(t, err)
		require.NotNil(t, proof)
		if tc.nonmembership {
			require.Equal(t, tc.value, proof.GetExclusion().GetActualValueHash())
			require.NotNil(t, proof.GetExclusion().GetLeaf())
			require.NotNil(t, proof.GetExclusion().GetPath())
		} else {
			require.Equal(t, tc.value, proof.GetExist().GetValue())
			require.NotNil(t, proof.GetExist().GetLeaf())
			require.NotNil(t, proof.GetExist().GetPath())
		}
	}

	err = nodeStore.Stop()
	require.NoError(t, err)
}

func TestICS23Proofs_VerifyCommitmentProofs(t *testing.T) {
	nodeStore := kvstore.NewMemKVStore()
	tree := smt.NewSparseMerkleTree(nodeStore, sha256.New(), smt.WithValueHasher(nil))
	require.NotNil(t, tree)

	// Set a value in the store
	err := tree.Update([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	err = tree.Update([]byte("bar"), []byte("foo"))
	require.NoError(t, err)
	err = tree.Update([]byte("testKey"), []byte("testValue"))
	require.NoError(t, err)
	err = tree.Update([]byte("testKey2"), []byte("testValue2"))
	require.NoError(t, err)

	root := tree.Root()
	require.NotNil(t, root)

	testCases := []struct {
		key           []byte
		value         []byte
		nonmembership bool
		valid         bool
	}{
		{
			// Successfully verifies a membership proof for a key-value stored pair
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         true,
		},
		{
			// Successfully verifies a non-membership proof for a key-value pair not stored
			key:           []byte("not stored"),
			value:         nil,
			nonmembership: true,
			valid:         true,
		},
		{
			// Fails to verify a membership proof for a key-value pair not stored
			key:           []byte("baz"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         false,
		},
		{
			// Fails to verify a non-membership proof for a key stored in the tree
			key:           []byte("foo"),
			value:         nil,
			nonmembership: true,
			valid:         false,
		},
	}

	proof := new(ics23.CommitmentProof)
	for _, tc := range testCases {
		var err error
		if tc.nonmembership {
			proof, err = createNonMembershipProof(tree, tc.key)
		} else {
			proof, err = createMembershipProof(tree, tc.key, tc.value)
		}
		require.NoError(t, err)
		require.NotNil(t, proof)

		if tc.nonmembership {
			require.NotNil(t, proof.GetExclusion())
		} else {
			require.NotNil(t, proof.GetExist())
		}

		var valid bool
		if tc.nonmembership {
			valid = VerifyNonMembership(root, proof, tc.key)
		} else {
			valid = VerifyMembership(root, proof, tc.key, tc.value)
		}

		if tc.valid {
			require.True(t, valid)
		} else {
			require.False(t, valid)
		}
	}

	err = nodeStore.Stop()
	require.NoError(t, err)
}
