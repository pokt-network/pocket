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

func TestTreeStore_AtomicUpdatesWithSuccessfulRollback(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
	mockBus := mockModules.NewMockBus(ctrl)
	mockPersistenceMod := mockModules.NewMockPersistenceModule(ctrl)

	mockBus.EXPECT().GetPersistenceModule().AnyTimes().Return(mockPersistenceMod)
	mockPersistenceMod.EXPECT().GetTxIndexer().AnyTimes().Return(mockTxIndexer)

	ts := &treeStore{
		logger:       logger.Global.CreateLoggerForModule(modules.TreeStoreSubmoduleName), // TODO
		treeStoreDir: ":memory:",
	}
	require.NoError(t, ts.setupTrees())
	require.NotEmpty(t, ts.merkleTrees[TransactionsTreeName])

	hash0 := ts.getStateHash()
	require.NotEmpty(t, hash0)

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

	// set a new savepoint
	require.NoError(t, ts.Savepoint())
	require.NotEmpty(t, ts.PrevState.MerkleTrees)
	require.NotEmpty(t, ts.PrevState.RootTree)
	// assert that savepoint creation doesn't mutate state hash
	require.Equal(t, hash1, hex.EncodeToString(ts.PrevState.RootTree.tree.Root()))

	hash2 := ts.getStateHash()
	require.Equal(t, hash2, hash1)

	// validate that state tree was updated and a previous savepoint is created
	for _, treeName := range stateTreeNames {
		require.NotEmpty(t, ts.merkleTrees[treeName])
		require.NotEmpty(t, ts.PrevState.MerkleTrees[treeName])
	}

	// insert additional test data into all of the trees
	for _, treeName := range stateTreeNames {
		require.NoError(t, ts.merkleTrees[treeName].tree.Update([]byte("fiz"), []byte("buz")))
	}

	// rollback the changes made to the trees above BEFORE anything was committed
	ts.Rollback()

	// validate that the state hash is unchanged after new data was inserted but rolled back before commitment
	hash3 := ts.getStateHash()
	require.Equal(t, hash3, hash2)
}
