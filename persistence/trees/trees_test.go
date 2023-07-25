package trees

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

// TECHDEBT(#836): Tests added in https://github.com/pokt-network/pocket/pull/836
func TestTreeStore_Update(t *testing.T) {
	// TODO: Write test case for the Update method
	t.Skip("TODO: Write test case for Update method")
}

func TestTreeStore_New(t *testing.T) {
	// TODO: Write test case for the NewStateTrees function
	t.Skip("TODO: Write test case for NewStateTrees function")
}

func TestTreeStore_DebugClearAll(t *testing.T) {
	// TODO: Write test case for the DebugClearAll method
	t.Skip("TODO: Write test case for DebugClearAll method")
}

// TODO_AFTER(#861): Implement this test with the test suite available in #861
func TestTreeStore_GetTreeHashes(t *testing.T) {
	t.Skip("TODO: Write test case for GetTreeHashes method") // context: https://github.com/pokt-network/pocket/pull/915#discussion_r1267313664
}

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
		expectedErr error
	}{
		{
			name:        "valid proof: key and value in tree",
			treeName:    "test",
			key:         []byte("key"),
			value:       []byte("value"),
			expectedErr: nil,
		},
		{
			name:        "valid proof: key not in tree",
			treeName:    "test",
			key:         []byte("key2"),
			value:       nil,
			expectedErr: nil,
		},
		{
			name:        "invalid proof: tree not in store",
			treeName:    "unstored tree",
			key:         []byte("key"),
			value:       []byte("value"),
			expectedErr: fmt.Errorf("tree not found: %s", "unstored tree"),
		},
		{
			name:        "invalid proof: key in tree, wrong value",
			treeName:    "test",
			key:         []byte("key"),
			value:       []byte("wrong value"),
			expectedErr: fmt.Errorf("invalid proof for key: %s, value: %s (%s)", hex.EncodeToString([]byte("key")), hex.EncodeToString([]byte("wrong value")), "test"),
		},
		{
			name:        "invalid proof: key in tree",
			treeName:    "test",
			key:         []byte("key"),
			value:       nil,
			expectedErr: fmt.Errorf("invalid proof for key: %s, value: %s (%s)", hex.EncodeToString([]byte("key")), hex.EncodeToString([]byte(nil)), "test"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := treeStore.Prove(tc.treeName, tc.key, tc.value)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorAs(t, err, &tc.expectedErr)
		})
	}
}
