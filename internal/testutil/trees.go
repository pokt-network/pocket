package testutil

import (
	"log"
	"testing"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/trees"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"

	"github.com/stretchr/testify/require"
)

var (
	testSchema = "test_schema"

	genesisStateNumValidators   = 5
	genesisStateNumServicers    = 1
	genesisStateNumApplications = 1
)

// creates a new tree store with a tmp directory for nodestore persistence
// and then starts the tree store and returns its pointer.
func NewTestTreeStoreSubmodule(t *testing.T, bus modules.Bus) modules.TreeStoreModule {
	t.Helper()

	tmpDir := t.TempDir()
	ts, err := trees.Create(
		bus,
		trees.WithTreeStoreDirectory(tmpDir),
		trees.WithLogger(logger.Global.CreateLoggerForModule(modules.TreeStoreSubmoduleName)))
	require.NoError(t, err)

	err = ts.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		err := ts.Stop()
		require.NoError(t, err)
	})

	return ts
}

func SeedTestTreeStoreSubmodule(t *testing.T, mod modules.TreeStoreModule) modules.TreeStoreModule {
	// TODO insert transaction data into postgres
	// TODO trigger an update with a pgx connection
	return mod
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
	require.NoError(t, err)

	persistenceMod, err := persistence.Create(bus)
	require.NoError(t, err)

	return persistenceMod.(modules.PersistenceModule)
}

func NewTestPostgresContext(t testing.TB, pmod modules.PersistenceModule, height int64) *persistence.PostgresContext {
	rwCtx, err := pmod.NewRWContext(height)
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
		resetStateToGenesis(pmod)
	})

	return postgresCtx
}

// This is necessary for unit tests that are dependant on a baseline genesis state
func resetStateToGenesis(pmod modules.PersistenceModule) {
	if err := pmod.ReleaseWriteContext(); err != nil {
		log.Fatalf("Error releasing write context: %v\n", err)
	}
	if err := pmod.HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	}); err != nil {
		log.Fatalf("Error clearing state: %v\n", err)
	}
}
