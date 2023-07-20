package trees

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/logger"
	mock_types "github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const (
	// the root hash of a tree store where each tree is empty but present and initialized
	h0 = "302f2956c084cc3e0e760cf1b8c2da5de79c45fa542f68a660a5fc494b486972"
	// the root hash of a tree store where each tree has has key foo value bar added to it
	h1 = "7d5712ea1507915c40e295845fa58773baa405b24b87e9d99761125d826ff915"
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

	for _, treeName := range stateTreeNames {
		fmt.Printf("%s: %s\n", treeName, hex.EncodeToString(ts.merkleTrees[treeName].tree.Root()))
	}

	// rollback the changes made to the trees above BEFORE anything was committed
	err := ts.Rollback()
	require.NoError(t, err)

	// validate that the state hash is unchanged after new data was inserted but rolled back before commitment
	hash3 := ts.getStateHash()
	require.Equal(t, hash3, hash2)
	require.Equal(t, hash3, h1)
	ts.Rollback()

	// confirm it's not in the tree
	v, err := ts.merkleTrees[TransactionsTreeName].tree.Get([]byte("fiz"))
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestTreeStore_SaveAndLoad(t *testing.T) {
	ctrl := gomock.NewController(t)
	tmpDir := t.TempDir()

	mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
	mockBus := mockModules.NewMockBus(ctrl)
	mockPersistenceMod := mockModules.NewMockPersistenceModule(ctrl)

	mockBus.EXPECT().GetPersistenceModule().AnyTimes().Return(mockPersistenceMod)
	mockPersistenceMod.EXPECT().GetTxIndexer().AnyTimes().Return(mockTxIndexer)

	ts := &treeStore{
		logger:       &zerolog.Logger{},
		treeStoreDir: tmpDir,
	}
	require.NoError(t, ts.Start())
	require.NotNil(t, ts.rootTree.tree)

	for _, treeName := range stateTreeNames {
		err := ts.merkleTrees[treeName].tree.Update([]byte("foo"), []byte("bar"))
		require.NoError(t, err)
	}

	err := ts.Commit()
	require.NoError(t, err)

	hash1 := ts.getStateHash()
	require.NotEmpty(t, hash1)

	w, err := ts.save()
	require.NoError(t, err)
	require.NotNil(t, w)
	require.NotNil(t, w.rootHash)
	require.NotNil(t, w.merkleRoots)

	// Stop the first tree store so that it's databases are no longer used
	require.NoError(t, ts.Stop())

	// declare a second TreeStore with no trees then load the first worldstate into it
	ts2 := &treeStore{
		logger:       logger.Global.CreateLoggerForModule(modules.TreeStoreSubmoduleName),
		treeStoreDir: tmpDir,
	}
	// TODO IN THIS COMMIT do we need to start this treestore?
	// require.NoError(t, ts2.Start())

	// Load sets a tree store to the provided worldstate
	err = ts2.Load(w)
	require.NoError(t, err)

	hash2 := ts2.getStateHash()

	// Assert that hash is unchanged from save and load
	require.Equal(t, hash1, hash2)
}
