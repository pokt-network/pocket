package trees

import (
	"encoding/hex"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/logger"
	mock_types "github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"

	"github.com/stretchr/testify/require"
)

const (
	// the root hash of a tree store where each tree is empty but present and initialized
	h0 = "8ba9f76843b2873557dee21b4894e9b308c2717f9c4e69a871ef085a193f842d"
	// the root hash of a tree store where each tree has has key foo value bar added to it
	h1 = "2096fdc67bdbe59cf3918f08203b6cf49c8c4ccee7a77906a434f9a2272826f4"
)

func TestTreeStore_AtomicUpdatesWithSuccessfulRollback(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
	mockBus := mockModules.NewMockBus(ctrl)
	mockPersistenceMod := mockModules.NewMockPersistenceModule(ctrl)

	mockBus.EXPECT().GetPersistenceModule().AnyTimes().Return(mockPersistenceMod)
	mockPersistenceMod.EXPECT().GetTxIndexer().AnyTimes().Return(mockTxIndexer)

	ts := &treeStore{
		logger:       logger.Global.CreateLoggerForModule(modules.TreeStoreSubmoduleName),
		treeStoreDir: ":memory:",
	}
	require.NoError(t, ts.setupTrees())
	require.NotEmpty(t, ts.merkleTrees[TransactionsTreeName])

	hash0 := ts.getStateHash()
	require.NotEmpty(t, hash0)
	require.Equal(t, hash0, h0)

	require.NoError(t, ts.Savepoint())

	// insert test data into every tree
	for _, treeName := range stateTreeNames {
		err := ts.merkleTrees[treeName].tree.Update([]byte("foo"), []byte("bar"))
		require.NoError(t, err)
	}

	// commit the above changes
	require.NoError(t, ts.Commit())

	// assert state hash is changed
	hash1 := ts.getStateHash()
	require.NotEmpty(t, hash1)
	require.NotEqual(t, hash0, hash1)
	require.Equal(t, hash1, h1)

	// set a new savepoint
	require.NoError(t, ts.Savepoint())
	require.NotEmpty(t, ts.prevState.merkleTrees)
	require.NotEmpty(t, ts.prevState.rootTree)
	// assert that savepoint creation doesn't mutate state hash
	require.Equal(t, hash1, hex.EncodeToString(ts.prevState.rootTree.tree.Root()))

	// verify that creating a savepoint does not change state hash
	hash2 := ts.getStateHash()
	require.Equal(t, hash2, hash1)
	require.Equal(t, hash2, h1)

	// validate that state tree was updated and a previous savepoint is created
	for _, treeName := range stateTreeNames {
		require.NotEmpty(t, ts.merkleTrees[treeName])
		require.NotEmpty(t, ts.prevState.merkleTrees[treeName])
	}

	// insert additional test data into all of the trees
	for _, treeName := range stateTreeNames {
		require.NoError(t, ts.merkleTrees[treeName].tree.Update([]byte("fiz"), []byte("buz")))
	}

	// rollback the changes made to the trees above BEFORE anything was committed
	err := ts.Rollback()
	require.NoError(t, err)

	// validate that the state hash is unchanged after new data was inserted but rolled back before commitment
	hash3 := ts.getStateHash()
	require.Equal(t, hash3, hash2)
	require.Equal(t, hash3, h1)
}
