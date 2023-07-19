package trees

import (
	"encoding/hex"
	"testing"

	"github.com/golang/mock/gomock"
	mock_types "github.com/pokt-network/pocket/persistence/types/mocks"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/rs/zerolog"

	"github.com/stretchr/testify/require"
)

func TestTreeStore_AtomicUpdates(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
	mockBus := mockModules.NewMockBus(ctrl)
	mockPersistenceMod := mockModules.NewMockPersistenceModule(ctrl)

	mockBus.EXPECT().GetPersistenceModule().AnyTimes().Return(mockPersistenceMod)
	mockPersistenceMod.EXPECT().GetTxIndexer().AnyTimes().Return(mockTxIndexer)

	ts := &treeStore{
		logger:       &zerolog.Logger{},
		treeStoreDir: ":memory:",
	}
	require.NoError(t, ts.setupTrees())
	require.NotEmpty(t, ts.merkleTrees[TransactionsTreeName])

	hash0 := ts.getStateHash()
	require.NotEmpty(t, hash0)

	require.NoError(t, ts.Savepoint())

	for _, treeName := range stateTreeNames {
		err := ts.merkleTrees[treeName].tree.Update([]byte("foo"), []byte("bar"))
		require.NoError(t, err)
	}

	require.NoError(t, ts.Commit())

	hash1 := ts.getStateHash()
	require.NotEmpty(t, hash1)
	require.NotEqual(t, hash0, hash1)

	require.NoError(t, ts.Savepoint())
	require.NotEmpty(t, ts.Prev.MerkleTrees)
	require.NotEmpty(t, ts.Prev.RootTree)
	require.Equal(t, hash1, hex.EncodeToString(ts.Prev.RootTree.tree.Root()))

	hash2 := ts.getStateHash()
	require.Equal(t, hash2, hash1)

	for _, treeName := range stateTreeNames {
		require.NotEmpty(t, ts.merkleTrees[treeName])
		require.NotEmpty(t, ts.Prev.MerkleTrees[treeName])
	}

	for _, treeName := range stateTreeNames {
		require.NoError(t, ts.merkleTrees[treeName].tree.Update([]byte("fiz"), []byte("buz")))
	}

	ts.Rollback()

	hash3 := ts.getStateHash()
	require.Equal(t, hash3, hash2)
}
