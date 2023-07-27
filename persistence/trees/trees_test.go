package trees_test

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/trees"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/utils"

	"github.com/stretchr/testify/require"
)

var (
	defaultChains          = []string{"0001"}
	defaultStakeBig        = big.NewInt(1000000000000000)
	defaultStake           = utils.BigIntToString(defaultStakeBig)
	defaultStakeStatus     = int32(coreTypes.StakeStatus_Staked)
	defaultPauseHeight     = int64(-1) // pauseHeight=-1 implies not paused
	defaultUnstakingHeight = int64(-1) // unstakingHeight=-1 implies not unstaking

	testSchema = "test_schema"

	genesisStateNumValidators   = 5
	genesisStateNumServicers    = 1
	genesisStateNumApplications = 1
)

const (
	trees_hash1 = "5282ee91a3ec0a6f2b30e4780b369bae78c80ef3ea40587fef6ae263bf41f244"
)

func TestTreeStore_Update(t *testing.T) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	t.Cleanup(func() {
		require.NoError(t, pool.Purge(resource))
	})

	t.Run("should update actor trees, commit, and modify the state hash", func(t *testing.T) {
		pmod := newTestPersistenceModule(t, dbUrl)
		context := newTestPostgresContext(t, 0, pmod)

		require.NoError(t, context.SetSavePoint())

		hash1, err := context.ComputeStateHash()
		require.NoError(t, err)
		require.NotEmpty(t, hash1)
		require.Equal(t, hash1, trees_hash1)

		_, err = createAndInsertDefaultTestApp(t, context)
		require.NoError(t, err)

		require.NoError(t, context.SetSavePoint())

		hash2, err := context.ComputeStateHash()
		require.NoError(t, err)
		require.NotEmpty(t, hash2)
		require.NotEqual(t, hash1, hash2)
	})

	t.Run("should fail to rollback when no treestore savepoint is set", func(t *testing.T) {
		pmod := newTestPersistenceModule(t, dbUrl)
		context := newTestPostgresContext(t, 0, pmod)

		err := context.RollbackToSavePoint()
		require.Error(t, err)
		require.ErrorIs(t, err, trees.ErrFailedRollback)
	})
}

func newTestPersistenceModule(t *testing.T, databaseURL string) modules.PersistenceModule {
	t.Helper()
	teardownDeterministicKeygen := keygen.GetInstance().SetSeed(42)
	defer teardownDeterministicKeygen()

	cfg := newTestDefaultConfig(t, databaseURL)
	genesisState, _ := test_artifacts.NewGenesisState(
		genesisStateNumValidators,
		genesisStateNumServicers,
		genesisStateNumApplications,
		genesisStateNumServicers,
	)

	runtimeMgr := runtime.NewManager(cfg, genesisState)

	bus, err := runtime.CreateBus(runtimeMgr)
	require.NoError(t, err)

	persistenceMod, err := persistence.Create(bus)
	require.NoError(t, err)

	return persistenceMod.(modules.PersistenceModule)
}

// fetches a new default node configuration for testing
func newTestDefaultConfig(t *testing.T, databaseURL string) *configs.Config {
	t.Helper()
	cfg := &configs.Config{
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseURL,
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
	return cfg
}
func createAndInsertDefaultTestApp(t *testing.T, db *persistence.PostgresContext) (*coreTypes.Actor, error) {
	t.Helper()
	app := newTestApp(t)

	addrBz, err := hex.DecodeString(app.Address)
	require.NoError(t, err)

	pubKeyBz, err := hex.DecodeString(app.PublicKey)
	require.NoError(t, err)

	outputBz, err := hex.DecodeString(app.Output)
	require.NoError(t, err)
	return app, db.InsertApp(
		addrBz,
		pubKeyBz,
		outputBz,
		false,
		defaultStakeStatus,
		defaultStake,
		defaultChains,
		defaultPauseHeight,
		defaultUnstakingHeight)
}

// TECHDEBT(#796): Test helpers should be consolidated in a single place
func newTestApp(t *testing.T) *coreTypes.Actor {
	operatorKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	outputAddr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	return &coreTypes.Actor{
		Address:         hex.EncodeToString(operatorKey.Address()),
		PublicKey:       hex.EncodeToString(operatorKey.Bytes()),
		Chains:          defaultChains,
		StakedAmount:    defaultStake,
		PausedHeight:    defaultPauseHeight,
		UnstakingHeight: defaultUnstakingHeight,
		Output:          hex.EncodeToString(outputAddr),
	}
}

// TECHDEBT(#796): Test helpers should be consolidated in a single place
func newTestPostgresContext(t testing.TB, height int64, testPersistenceMod modules.PersistenceModule) *persistence.PostgresContext {
	t.Helper()
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
