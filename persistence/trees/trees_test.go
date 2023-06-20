package trees

import (
	"crypto/sha256"
	"testing"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

func TestTreeStore_Update(t *testing.T) {
	// TODO: Write test case for the Update method
	t.Skip("TODO: Write test case for Update method")
}

func TestNewStateTrees(t *testing.T) {
	// TODO: Write test case for the NewStateTrees function
	t.Skip("TODO: Write test case for NewStateTrees function")
}

func TestTreeStore_DebugClearAll(t *testing.T) {
	// TODO: Write test case for the DebugClearAll method
	t.Skip("TODO: Write test case for DebugClearAll method")
}

func TestStateTree_Operations(t *testing.T) {
	nodeStore := kvstore.NewMemKVStore()
	require.NotNil(t, nodeStore)
	tree := smt.NewSparseMerkleTree(nodeStore, sha256.New())
	require.NotNil(t, tree)

	stateTree := &StateTree{
		key:       []byte("test"),
		tree:      tree,
		nodeStore: nodeStore,
	}

	require.Equal(t, stateTree.GetKey(), []byte("test"))
	require.Equal(t, stateTree.GetTree(), tree)
	require.Equal(t, stateTree.GetNodeStore(), nodeStore)

	// insert values into tree
	stateTree.tree.Update([]byte("key1"), []byte("value1"))
	stateTree.tree.Update([]byte("key2"), []byte("value2"))
	stateTree.tree.Update([]byte("key3"), []byte("value3"))

	root := stateTree.tree.Root()
	require.Equal(t, root, stateTree.GetTree().Root())
}
