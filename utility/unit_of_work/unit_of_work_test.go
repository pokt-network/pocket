package unit_of_work

import (
	"log"
	"os"
	"testing"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	utilTypes "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

const (
	testingValidatorCount   = 5
	testingServicerCount    = 1
	testingApplicationCount = 1
	testingFishermenCount   = 1

	testNonce  = "defaultNonceString"
	testSchema = "test_schema"
)

var (
	// TODO(#261): Utility module tests should have no dependencies on the persistence module (which instantiates a postgres container)
	testPersistenceMod modules.PersistenceModule
)

func NewTestingMempool(_ *testing.T) mempool.TXMempool {
	return utilTypes.NewTxFIFOMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbURL := test_artifacts.SetupPostgresDocker()

	runtimeCfg := newTestRuntimeConfig(dbURL)
	bus, err := runtime.CreateBus(runtimeCfg)
	if err != nil {
		log.Fatalf("Error creating bus: %s", err)
	}

	testPersistenceMod = newTestPersistenceModule(bus)

	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func newTestingUtilityUnitOfWork(t *testing.T, height int64, options ...func(*baseUtilityUnitOfWork)) *baseUtilityUnitOfWork {
	rwCtx, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	// TECHDEBT: Move the internal of cleanup into a separate function and call this in the
	// beginning of every test. This (the current implementation) is an issue because if we call
	// `NewTestingUtilityContext` more than once in a single test, we create unnecessary calls to clean.
	t.Cleanup(func() {
		err := testPersistenceMod.ReleaseWriteContext()
		require.NoError(t, err)
		err = testPersistenceMod.HandleDebugMessage(&messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
			Message: nil,
		})
		require.NoError(t, err)
		// TODO: May need to run `bus.GetUtilityModule().GetMempool().Clear()` here
	})

	uow := &baseUtilityUnitOfWork{
		logger: logger.Global.CreateLoggerForModule(modules.UtilityModuleName),
		height: height,
		// TODO(@deblasis): Refactor this
		persistenceRWContext:   rwCtx,
		persistenceReadContext: rwCtx,
	}

	uow.SetBus(testPersistenceMod.GetBus())

	for _, option := range options {
		option(uow)
	}

	return uow
}

func newTestPersistenceModule(bus modules.Bus) modules.PersistenceModule {
	persistenceMod, err := persistence.Create(bus)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod.(modules.PersistenceModule)
}

func newTestRuntimeConfig(databaseURL string) *runtime.Manager {
	cfg := &configs.Config{
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
		},
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseURL,
			NodeSchema:        testSchema,
			BlockStorePath:    "", // in memory
			TxIndexerPath:     "", // in memory
			TreesStoreDir:     "", // in memory
			MaxConnsCount:     50,
			MinConnsCount:     1,
			MaxConnLifetime:   "5m",
			MaxConnIdleTime:   "1m",
			HealthCheckPeriod: "30s",
		},
	}
	genesisState, _ := test_artifacts.NewGenesisState(
		testingValidatorCount,
		testingServicerCount,
		testingApplicationCount,
		testingFishermenCount,
	)
	runtimeCfg := runtime.NewManager(cfg, genesisState)
	return runtimeCfg
}
