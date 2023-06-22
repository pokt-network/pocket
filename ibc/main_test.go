package ibc

import (
	"log"
	"os"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
)

var dbURL string

// NB: `TestMain` serves all tests in the immediate `ibc` package and not its children
func TestMain(m *testing.M) {
	pool, resource, url := test_artifacts.SetupPostgresDocker()
	dbURL = url

	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func newTestUtilityModule(bus modules.Bus) modules.UtilityModule {
	utilityMod, err := utility.Create(bus)
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

func newTestIbcModule(bus modules.Bus) modules.IBCModule {
	ibcMod, err := Create(bus)
	if err != nil {
		log.Fatalf("Error creating ibc module: %s", err)
	}
	return ibcMod.(modules.IBCModule)
}

// Prepares a runtime environment for testing along with a genesis state, a persistence module and a utility module
//
//nolint:unparam // Test suite is not fully built out yet
func prepareEnvironment(
	t *testing.T,
	numValidators, // nolint:unparam // we are not currently modifying parameter but want to keep it modifiable in the future
	numServicers,
	numApplications,
	numFisherman int,
	genesisOpts ...test_artifacts.GenesisOption,
) (*runtime.Manager, modules.UtilityModule, modules.PersistenceModule, modules.IBCModule) {
	teardownDeterministicKeygen := keygen.GetInstance().SetSeed(42)

	runtimeCfg := newTestRuntimeConfig(numValidators, numServicers, numApplications, numFisherman, genesisOpts...)
	bus, err := runtime.CreateBus(runtimeCfg)
	require.NoError(t, err)

	testPersistenceMod := newTestPersistenceModule(bus)
	err = testPersistenceMod.Start()
	require.NoError(t, err)

	testUtilityMod := newTestUtilityModule(bus)
	err = testUtilityMod.Start()
	require.NoError(t, err)

	testIbcMod := newTestIbcModule(bus)
	err = testIbcMod.Start()
	require.NoError(t, err)

	// Reset database to genesis before every test
	err = testPersistenceMod.HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		teardownDeterministicKeygen()
		err := testPersistenceMod.Stop()
		require.NoError(t, err)
		err = testUtilityMod.Stop()
		require.NoError(t, err)
		err = testIbcMod.Stop()
		require.NoError(t, err)
	})

	return runtimeCfg, testUtilityMod, testPersistenceMod, testIbcMod
}

//nolint:unparam // Test suite is not fully built out yet
func newTestRuntimeConfig(
	numValidators,
	numServicers,
	numApplications,
	numFisherman int,
	genesisOpts ...test_artifacts.GenesisOption,
) *runtime.Manager {
	cfg, err := configs.CreateTempConfig(&configs.Config{
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
		},
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       dbURL,
			NodeSchema:        "test_schema",
			BlockStorePath:    ":memory:",
			TxIndexerPath:     ":memory:",
			TreesStoreDir:     ":memory:",
			MaxConnsCount:     50,
			MinConnsCount:     1,
			MaxConnLifetime:   "5m",
			MaxConnIdleTime:   "1m",
			HealthCheckPeriod: "30s",
		},
		Validator: &configs.ValidatorConfig{Enabled: true},
		IBC: &configs.IBCConfig{
			Enabled:    true,
			PrivateKey: "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
		},
	})
	if err != nil {
		log.Fatalf("Error creating config: %s", err)
	}
	genesisState, _ := test_artifacts.NewGenesisState(
		numValidators,
		numServicers,
		numApplications,
		numFisherman,
		genesisOpts...,
	)
	runtimeCfg := runtime.NewManager(cfg, genesisState)
	return runtimeCfg
}
