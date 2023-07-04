package ibc

import (
	"log"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
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

func newTestConsensusModule(bus modules.Bus) modules.ConsensusModule {
	consensusMod, err := consensus.Create(bus)
	if err != nil {
		log.Fatalf("Error creating consensus module: %s", err)
	}
	return consensusMod.(modules.ConsensusModule)
}

func newTestP2PModule(t *testing.T, bus modules.Bus) modules.P2PModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	p2pMock := mockModules.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(msg *anypb.Any) {
		}).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any) {
		}).
		AnyTimes()
	p2pMock.EXPECT().GetModuleName().Return(modules.P2PModuleName).AnyTimes()
	p2pMock.EXPECT().HandleEvent(gomock.Any()).Return(nil).AnyTimes()
	bus.RegisterModule(p2pMock)

	return p2pMock
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

func newTestIBCModule(bus modules.Bus) modules.IBCModule {
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
) (*runtime.Manager, modules.ConsensusModule, modules.UtilityModule, modules.PersistenceModule, modules.IBCModule) {
	teardownDeterministicKeygen := keygen.GetInstance().SetSeed(42)

	runtimeCfg := newTestRuntimeConfig(numValidators, numServicers, numApplications, numFisherman, genesisOpts...)
	bus, err := runtime.CreateBus(runtimeCfg)
	require.NoError(t, err)

	testPersistenceMod := newTestPersistenceModule(bus)
	err = testPersistenceMod.Start()
	require.NoError(t, err)

	testConsensusMod := newTestConsensusModule(bus)
	err = testConsensusMod.Start()
	require.NoError(t, err)

	testP2PMock := newTestP2PModule(t, bus)
	err = testP2PMock.Start()
	require.NoError(t, err)

	testUtilityMod := newTestUtilityModule(bus)
	err = testUtilityMod.Start()
	require.NoError(t, err)

	testIBCMod := newTestIBCModule(bus)
	err = testIBCMod.Start()
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
		err = testIBCMod.Stop()
		require.NoError(t, err)
	})

	return runtimeCfg, testConsensusMod, testUtilityMod, testPersistenceMod, testIBCMod
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
		Consensus: &configs.ConsensusConfig{
			PrivateKey: "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
		},
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
			Enabled: true,
			Host: &configs.IBCHostConfig{
				PrivateKey: "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
				StoresDir:  ":memory:",
			},
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
