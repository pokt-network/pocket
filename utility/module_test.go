package utility

import (
	"log"
	"os"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

var (
	dbURL string
)

func TestMain(m *testing.M) {
	pool, resource, url := test_artifacts.SetupPostgresDocker()
	dbURL = url

	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func newTestUtilityModule(bus modules.Bus) modules.UtilityModule {
	utilityMod, err := Create(bus)
	if err != nil {
		log.Fatalf("Error creating utility module: %s", err)
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

func prepareEnvironment(
	t *testing.T,
	numValidators,
	numServicers,
	numApplications,
	numFisherman int,
) (*runtime.Manager, modules.UtilityModule, modules.PersistenceModule) {
	teardownDeterministicKeygen := keygen.GetInstance().SetSeed(42)

	runtimeCfg := newTestRuntimeConfig(numValidators, numServicers, numApplications, numFisherman)
	bus, err := runtime.CreateBus(runtimeCfg)
	require.NoError(t, err)

	testPersistenceMod := newTestPersistenceModule(bus)
	testPersistenceMod.Start()

	testUtilityMod := newTestUtilityModule(bus)
	testUtilityMod.Start()

	t.Cleanup(func() {
		teardownDeterministicKeygen()
		testPersistenceMod.Stop()
		testUtilityMod.Stop()
	})

	return runtimeCfg, testUtilityMod, testPersistenceMod
}

// REFACTOR: This should be in a shared testing package
func newTestRuntimeConfig(
	numValidators,
	numServicers,
	numApplications,
	numFisherman int,
) *runtime.Manager {
	cfg := &configs.Config{
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
		},
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       dbURL,
			NodeSchema:        "test_schema",
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
		numValidators,
		numServicers,
		numApplications,
		numFisherman,
	)
	runtimeCfg := runtime.NewManager(cfg, genesisState)
	return runtimeCfg
}
