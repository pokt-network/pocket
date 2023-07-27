package trees

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

func TestTreeStore_Prove(t *testing.T) {
	nodeStore := kvstore.NewMemKVStore()
	tree := smt.NewSparseMerkleTree(nodeStore, smtTreeHasher)
	testTree := &stateTree{
		name:      "test",
		tree:      tree,
		nodeStore: nodeStore,
	}

	require.NoError(t, testTree.tree.Update([]byte("key"), []byte("value")))
	require.NoError(t, testTree.tree.Commit())

	treeStore := &treeStore{
		merkleTrees: make(map[string]*stateTree, 1),
	}
	treeStore.merkleTrees["test"] = testTree

	testCases := []struct {
		name        string
		treeName    string
		key         []byte
		value       []byte
		valid       bool
		expectedErr error
	}{
		{
			name:        "valid inclusion proof: key and value in tree",
			treeName:    "test",
			key:         []byte("key"),
			value:       []byte("value"),
			valid:       true,
			expectedErr: nil,
		},
		{
			name:        "valid exclusion proof: key not in tree",
			treeName:    "test",
			key:         []byte("key2"),
			value:       nil,
			valid:       true,
			expectedErr: nil,
		},
		{
			name:        "invalid proof: tree not in store",
			treeName:    "unstored tree",
			key:         []byte("key"),
			value:       []byte("value"),
			valid:       false,
			expectedErr: fmt.Errorf("tree not found: %s", "unstored tree"),
		},
		{
			name:        "invalid inclusion proof: key in tree, wrong value",
			treeName:    "test",
			key:         []byte("key"),
			value:       []byte("wrong value"),
			valid:       false,
			expectedErr: nil,
		},
		{
			name:        "invalid exclusion proof: key in tree",
			treeName:    "test",
			key:         []byte("key"),
			value:       nil,
			valid:       false,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := treeStore.Prove(tc.treeName, tc.key, tc.value)
			require.Equal(t, valid, tc.valid)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorAs(t, err, &tc.expectedErr)
		})
	}
}
