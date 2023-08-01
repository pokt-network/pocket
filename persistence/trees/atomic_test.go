package trees

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/pokt-network/pocket/logger"
	mock_types "github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const (
	// the root hash of a tree store where each tree is empty but present and initialized
	h0 = "302f2956c084cc3e0e760cf1b8c2da5de79c45fa542f68a660a5fc494b486972"
	// the root hash of a tree store where each tree has has key foo value bar added to it
	h1 = "7d5712ea1507915c40e295845fa58773baa405b24b87e9d99761125d826ff915"
)

var (
	testFoo = []byte("foo")
	testBar = []byte("bar")
	testKey = []byte("fiz")
	testVal = []byte("buz")
)

func TestTreeStore_AtomicUpdatesWithSuccessfulRollback(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
	mockBus := mock_modules.NewMockBus(ctrl)
	mockPersistenceMod := mock_modules.NewMockPersistenceModule(ctrl)

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
		err := ts.merkleTrees[treeName].tree.Update(testFoo, testBar)
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
		require.NoError(t, ts.merkleTrees[treeName].tree.Update(testKey, testVal))
	}

	// rollback the changes made to the trees above BEFORE anything was committed
	err := ts.Rollback()
	require.NoError(t, err)

	// validate that the state hash is unchanged after new data was inserted but rolled back before commitment
	hash3 := ts.getStateHash()
	require.Equal(t, hash3, hash2)
	require.Equal(t, hash3, h1)

	err = ts.Rollback()
	require.NoError(t, err)

	// confirm it's not in the tree
	v, err := ts.merkleTrees[TransactionsTreeName].tree.Get(testKey)
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestTreeStore_SaveAndLoad(t *testing.T) {
	t.Parallel()
	t.Run("should save a backup in a directory", func(t *testing.T) {
		ts := newTestTreeStore(t)
		backupDir := t.TempDir()
		// assert that the directory is empty before backup
		ok, err := isEmpty(backupDir)
		require.NoError(t, err)
		require.True(t, ok)

		// Trigger a backup
		require.NoError(t, ts.Backup(backupDir))

		// assert that the directory is not empty after Backup has returned
		ok, err = isEmpty(backupDir)
		require.NoError(t, err)
		require.False(t, ok)

		// assert that the worldstate.json file exists after a backup

		// Open the directory
		dir, err := os.Open(backupDir)
		if err != nil {
			fmt.Printf("Error opening directory: %s\n", err)
			return
		}
		defer dir.Close()

		// Read directory entries one by one
		files, err := dir.Readdir(0) // 0 means read all directory entries
		if err != nil {
			fmt.Printf("Error reading directory entries: %s\n", err)
			return
		}
		require.Equal(t, len(files), len(stateTreeNames)+1) // +1 to account for the worldstate file

		// Now files is a slice of FileInfo objects representing the directory entries
		// You can work with them as needed.
		for _, file := range files {
			if file.IsDir() {
				fmt.Printf("Directory: %s\n", file.Name())
			} else {
				fmt.Printf("File: %s\n", file.Name())
			}
		}
	})
	t.Run("should load a backup and maintain TreeStore hash integrity", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
		mockBus := mock_modules.NewMockBus(ctrl)
		mockPersistenceMod := mock_modules.NewMockPersistenceModule(ctrl)

		mockBus.EXPECT().GetPersistenceModule().AnyTimes().Return(mockPersistenceMod)
		mockPersistenceMod.EXPECT().GetTxIndexer().AnyTimes().Return(mockTxIndexer)

		// create a new tree store and save it's initial hash
		ts := newTestTreeStore(t)
		hash1 := ts.getStateHash()

		// make a temp directory for the backup and assert it's empty
		backupDir := t.TempDir()
		empty, err := isEmpty(backupDir)
		require.NoError(t, err)
		require.True(t, empty)

		// make a backup
		err = ts.Backup(backupDir)
		require.NoError(t, err)

		// assert directory is not empty after backup
		empty2, err := isEmpty(backupDir)
		require.NoError(t, err)
		require.False(t, empty2)

		// stop the first tree store so that it's databases are released
		require.NoError(t, ts.Stop())

		// declare a second TreeStore with no trees then load the first worldstate into it
		ts2 := &treeStore{
			logger: logger.Global.CreateLoggerForModule(modules.TreeStoreSubmoduleName),
		}

		// call load with the backup directory
		err = ts2.Load(backupDir)
		require.NoError(t, err)

		// Assert that hash is unchanged from save and load
		hash2 := ts2.getStateHash()
		require.Equal(t, hash1, hash2)
	})
}

// creates a new tree store with a tmp directory for nodestore persistence
// and then starts the tree store and returns its pointer.
// TECHDEBT(#796) - Organize and dedupe this function into testutil package
func newTestTreeStore(t *testing.T) *treeStore {
	t.Helper()
	ctrl := gomock.NewController(t)
	tmpDir := t.TempDir()

	mockTxIndexer := mock_types.NewMockTxIndexer(ctrl)
	mockBus := mock_modules.NewMockBus(ctrl)
	mockPersistenceMod := mock_modules.NewMockPersistenceModule(ctrl)

	mockBus.EXPECT().GetPersistenceModule().AnyTimes().Return(mockPersistenceMod)
	mockPersistenceMod.EXPECT().GetTxIndexer().AnyTimes().Return(mockTxIndexer)

	ts := &treeStore{
		logger:       logger.Global.CreateLoggerForModule(modules.TreeStoreSubmoduleName),
		treeStoreDir: tmpDir,
	}
	require.NoError(t, ts.Start())
	require.NotNil(t, ts.rootTree.tree)

	for _, treeName := range stateTreeNames {
		err := ts.merkleTrees[treeName].tree.Update(testFoo, testBar)
		require.NoError(t, err)
	}

	err := ts.Commit()
	require.NoError(t, err)

	hash1 := ts.getStateHash()
	require.NotEmpty(t, hash1)

	return ts
}

// TECHDEBT(#796) - Organize and dedupe this function into testutil package
func isEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
