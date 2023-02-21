package utility

import (
	"log"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
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
	// Initialized in TestMain
	testPersistenceMod modules.PersistenceModule
	testUtilityMod     modules.UtilityModule
)

func NewTestingMempool(_ *testing.T) mempool.TXMempool {
	return utilTypes.NewTxFIFOMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	// TODO(#261): Utility module tests should have no dependencies on a postgres container
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

	uc := &utilityContext{
		height:             height,
		persistenceContext: persistenceContext,
		savePointsSet:      make(map[string]struct{}),
		savePointsList:     make([][]byte, 0),
	}

	return uc.setBus(testUtilityMod.GetBus())
}

func newTestRuntimeConfig(databaseUrl string) *runtime.Manager {
	cfg := &configs.Config{
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       databaseUrl,
			NodeSchema:        testSchema,
			BlockStorePath:    "",
			TxIndexerPath:     "",
			TreesStoreDir:     "",
			MaxConnsCount:     4,
			MinConnsCount:     0,
			MaxConnLifetime:   "1h",
			MaxConnIdleTime:   "30m",
			HealthCheckPeriod: "5m",
		},
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
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

func newTestUtilityModule(bus modules.Bus) modules.UtilityModule {
	utilityMod, err := Create(bus)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return utilityMod.(modules.UtilityModule)
}

// TODO(#261): Utility module tests should have no dependencies on the persistence module
func newTestPersistenceModule(bus modules.Bus) modules.PersistenceModule {
	persistenceMod, err := persistence.Create(bus)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod.(modules.PersistenceModule)
}

// IMPROVE: Not part of `TestMain` because a mock requires `testing.T` to be initialized.
// We are trying to only initialize one `testPersistenceModule` in all the tests, so when the
// utility module tests are no longer dependant on the persistence module explicitly, this
// can be improved.
func mockBusInTestModules(t *testing.T) {
	ctrl := gomock.NewController(t)

	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(testPersistenceMod).AnyTimes()
	busMock.EXPECT().GetUtilityModule().Return(testUtilityMod).AnyTimes()

	testPersistenceMod.SetBus(busMock)
	testUtilityMod.SetBus(busMock)

	t.Cleanup(func() {
		testPersistenceMod.SetBus(nil)
	})
}
