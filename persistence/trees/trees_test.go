package trees_test

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	DefaultChains = []string{"0001"}

	DefaultDeltaBig   = big.NewInt(100)
	DefaultAccountBig = big.NewInt(1000000)
	DefaultStakeBig   = big.NewInt(1000000000000000)

	DefaultAccountAmount = utils.BigIntToString(DefaultAccountBig)
	DefaultStake         = utils.BigIntToString(DefaultStakeBig)

	DefaultStakeStatus     = int32(coreTypes.StakeStatus_Staked)
	DefaultPauseHeight     = int64(-1) // pauseHeight=-1 implies not paused
	DefaultUnstakingHeight = int64(-1) // unstakingHeight=-1 implies not unstaking

	testSchema = "test_schema"

	genesisStateNumValidators   = 5
	genesisStateNumServicers    = 1
	genesisStateNumApplications = 1
)

func TestTreeStore_Update(t *testing.T) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("could not purge resource: %s", err)
		}
	})

	t.Run("should update actor tree and commit", func(t *testing.T) {
		pmod := newTestPersistenceModule(t, dbUrl)
		context := NewTestPostgresContext(t, 0, pmod)

		actor, err := createAndInsertDefaultTestApp(t, context)
		assert.NoError(t, err)

		t.Logf("actor inserted %+v", actor)
		hash, err := context.ComputeStateHash()
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

func newTestPersistenceModule(t *testing.T, databaseUrl string) modules.PersistenceModule {
	t.Helper()
	teardownDeterministicKeygen := keygen.GetInstance().SetSeed(42)
	defer teardownDeterministicKeygen()

	cfg := &configs.Config{
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseUrl,
			NodeSchema:        testSchema,
			BlockStorePath:    ":memory:",
			TxIndexerPath:     ":memory:",
			TreesStoreDir:     ":memory:",
			MaxConnsCount:     5,
			MinConnsCount:     1,
			MaxConnLifetime:   "5m",
			MaxConnIdleTime:   "1m",
			HealthCheckPeriod: "30s",
		},
	}

	genesisState, _ := test_artifacts.NewGenesisState(
		genesisStateNumValidators,
		genesisStateNumServicers,
		genesisStateNumApplications,
		genesisStateNumServicers,
	)

	runtimeMgr := runtime.NewManager(cfg, genesisState)

	bus, err := runtime.CreateBus(runtimeMgr)
	assert.NoError(t, err)

	persistenceMod, err := persistence.Create(bus)
	assert.NoError(t, err)

	return persistenceMod.(modules.PersistenceModule)
}
func createAndInsertDefaultTestApp(t *testing.T, db *persistence.PostgresContext) (*coreTypes.Actor, error) {
	app, err := newTestApp()
	if err != nil {
		return nil, err
	}
	// TODO(andrew): Avoid the use of `log.Fatal(fmt.Sprintf`
	// TODO(andrew): Use `require.NoError` instead of `log.Fatal` in tests`
	addrBz, err := hex.DecodeString(app.Address)
	if err != nil {
		t.Errorf("an error occurred converting address to bytes %s", app.Address)
	}
	pubKeyBz, err := hex.DecodeString(app.PublicKey)
	if err != nil {
		t.Errorf("an error occurred converting pubKey to bytes %s", app.PublicKey)
	}
	outputBz, err := hex.DecodeString(app.Output)
	if err != nil {
		t.Errorf("an error occurred converting output to bytes %s", app.Output)
	}
	return app, db.InsertApp(
		addrBz,
		pubKeyBz,
		outputBz,
		false,
		DefaultStakeStatus,
		DefaultStake,
		DefaultChains,
		DefaultPauseHeight,
		DefaultUnstakingHeight)
}

func newTestApp() (*coreTypes.Actor, error) {
	operatorKey, err := crypto.GeneratePublicKey()
	if err != nil {
		return nil, err
	}

	outputAddr, err := crypto.GenerateAddress()
	if err != nil {
		return nil, err
	}

	return &coreTypes.Actor{
		Address:         hex.EncodeToString(operatorKey.Address()),
		PublicKey:       hex.EncodeToString(operatorKey.Bytes()),
		Chains:          DefaultChains,
		StakedAmount:    DefaultStake,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}, nil
}

func NewTestPostgresContext(t testing.TB, height int64, testPersistenceMod modules.PersistenceModule) *persistence.PostgresContext {
	rwCtx, err := testPersistenceMod.NewRWContext(height)
	if err != nil {
		log.Fatalf("Error creating new context: %v\n", err)
	}

	postgresCtx, ok := rwCtx.(*persistence.PostgresContext)
	if !ok {
		log.Fatalf("Error casting RW context to Postgres context")
	}

	// TECHDEBT: This should not be part of `NewTestPostgresContext`. It causes unnecessary resets
	// if we call `NewTestPostgresContext` more than once in a single test.
	t.Cleanup(func() {
		resetStateToGenesis(testPersistenceMod)
	})

	return postgresCtx
}

func NewTestPersistenceModule(t *testing.T, databaseUrl string) modules.PersistenceModule {
	teardownDeterministicKeygen := keygen.GetInstance().SetSeed(42)
	defer teardownDeterministicKeygen()

	cfg := &configs.Config{
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseUrl,
			NodeSchema:        testSchema,
			BlockStorePath:    ":memory:",
			TxIndexerPath:     ":memory:",
			TreesStoreDir:     ":memory:",
			MaxConnsCount:     5,
			MinConnsCount:     1,
			MaxConnLifetime:   "5m",
			MaxConnIdleTime:   "1m",
			HealthCheckPeriod: "30s",
		},
	}

	genesisState, _ := test_artifacts.NewGenesisState(
		genesisStateNumValidators,
		genesisStateNumServicers,
		genesisStateNumApplications,
		genesisStateNumServicers,
	)
	runtimeMgr := runtime.NewManager(cfg, genesisState)
	bus, err := runtime.CreateBus(runtimeMgr)
	if err != nil {
		log.Printf("Error creating bus: %s", err)
		return nil
	}

	persistenceMod, err := persistence.Create(bus)
	if err != nil {
		log.Printf("Error creating persistence module: %s", err)
		return nil
	}

	return persistenceMod.(modules.PersistenceModule)
}

// This is necessary for unit tests that are dependant on a baseline genesis state
func resetStateToGenesis(m modules.PersistenceModule) {
	if err := m.ReleaseWriteContext(); err != nil {
		log.Fatalf("Error releasing write context: %v\n", err)
	}
	if err := m.HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	}); err != nil {
		log.Fatalf("Error clearing state: %v\n", err)
	}
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
