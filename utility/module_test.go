package utility

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
	testUtilityMod modules.UtilityModule

	// TODO(#261): Utility module tests should have no dependencies on the persistence module (which instantiates a postgres container)
	testPersistenceMod modules.PersistenceModule
)

func NewTestingMempool(_ *testing.T) mempool.TXMempool {
	return utilTypes.NewTxFIFOMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()

	runtimeCfg := newTestRuntimeConfig(dbUrl)
	bus, err := runtime.CreateBus(runtimeCfg)
	if err != nil {
		log.Fatalf("Error creating bus: %s", err)
	}

	testUtilityMod = newTestUtilityModule(bus)
	testPersistenceMod = newTestPersistenceModule(bus)

	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func newTestingUtilityContext(t *testing.T, height int64) *utilityContext {
	persistenceContext, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	// TECHDEBT: Move the internal of cleanup into a separate function and call this in the
	// beginning of every test. This (the current implementation) is an issue because if we call
	// `NewTestingUtilityContext` more than once in a single test, we create unnecessary calls to clean.
	t.Cleanup(func() {
		require.NoError(t, testPersistenceMod.ReleaseWriteContext())
		require.NoError(t, testPersistenceMod.HandleDebugMessage(&messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
			Message: nil,
		}))
		testUtilityMod.GetMempool().Clear()
	})

	ctx := &utilityContext{
		logger:         logger.Global.CreateLoggerForModule(modules.UtilityModuleName),
		height:         height,
		store:          persistenceContext,
		savePointsSet:  make(map[string]struct{}),
		savePointsList: make([][]byte, 0),
	}
	ctx.IntegratableModule.SetBus(testUtilityMod.GetBus())

	return ctx
}

func newTestUtilityModule(bus modules.Bus) modules.UtilityModule {
	utilityMod, err := Create(bus)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return utilityMod.(modules.UtilityModule)
}

func newTestPersistenceModule(bus modules.Bus) modules.PersistenceModule {
	persistenceMod, err := persistence.Create(bus)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod.(modules.PersistenceModule)
}

func newTestRuntimeConfig(databaseUrl string) *runtime.Manager {
	cfg := &configs.Config{
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
		},
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseUrl,
			NodeSchema:        testSchema,
			BlockStorePath:    "", // in memory
			TxIndexerPath:     "", // in memory
			TreesStoreDir:     "", // in memory
			MaxConnsCount:     4,
			MinConnsCount:     0,
			MaxConnLifetime:   "1h",
			MaxConnIdleTime:   "30m",
			HealthCheckPeriod: "5m",
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
