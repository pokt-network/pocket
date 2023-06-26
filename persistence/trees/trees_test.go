package trees_test

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/stretchr/testify/assert"
)

var (
	DefaultChains     = []string{"0001"}
	ChainsToUpdate    = []string{"0002"}
	DefaultServiceURL = "https://foo.bar"
	DefaultPoolName   = "TESTING_POOL"

	DefaultDeltaBig   = big.NewInt(100)
	DefaultAccountBig = big.NewInt(1000000)
	DefaultStakeBig   = big.NewInt(1000000000000000)

	DefaultAccountAmount = utils.BigIntToString(DefaultAccountBig)
	DefaultStake         = utils.BigIntToString(DefaultStakeBig)
	StakeToUpdate        = utils.BigIntToString((&big.Int{}).Add(DefaultStakeBig, DefaultDeltaBig))

	DefaultStakeStatus     = int32(coreTypes.StakeStatus_Staked)
	DefaultPauseHeight     = int64(-1) // pauseHeight=-1 implies not paused
	DefaultUnstakingHeight = int64(-1) // unstakingHeight=-1 implies not unstaking

	OlshanskyURL    = "https://olshansky.info"
	OlshanskyChains = []string{"OLSH"}

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

		actor, err := createAndInsertDefaultTestApp(context)
		assert.NoError(t, err)

		t.Logf("actor inserted %+v", actor)
		hash, err := context.ComputeStateHash()
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("should apply commit to all stores and return nil", func(t *testing.T) {
		pmod := newTestPersistenceModule(t, dbUrl)
		context := NewTestPostgresContext(t, 0, pmod)

		_, err := createAndInsertDefaultTestApp(context)
		assert.NoError(t, err)

		// create a block for testing
		testblock := &coreTypes.Block{
			BlockHeader: &coreTypes.BlockHeader{
				Height: 1,
			},
			Transactions: [][]byte{
				[]byte("hello"),
				[]byte("world"),
			},
		}

		// apply the transaction
		err = context.Apply(modules.Tx{
			Block: testblock,
		})
		assert.NoError(t, err)
	})

	t.Run("should rollback if any store fails to update", func(t *testing.T) {
		t.Skip()
	})
}

// createMockBus returns a mock bus with stubbed out functions for bus registration
func createMockBus(t *testing.T, runtimeMgr modules.RuntimeMgr) *mockModules.MockBus {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockBus := mockModules.NewMockBus(ctrl)
	mockBus.EXPECT().GetRuntimeMgr().Return(runtimeMgr).AnyTimes()
	mockBus.EXPECT().RegisterModule(gomock.Any()).DoAndReturn(func(m modules.Module) {
		m.SetBus(mockBus)
	}).AnyTimes()
	mockModulesRegistry := mockModules.NewMockModulesRegistry(ctrl)
	mockModulesRegistry.EXPECT().GetModule(peerstore_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(peerstore_provider.ModuleName)).AnyTimes()
	mockModulesRegistry.EXPECT().GetModule(current_height_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(current_height_provider.ModuleName)).AnyTimes()
	mockBus.EXPECT().GetModulesRegistry().Return(mockModulesRegistry).AnyTimes()
	mockBus.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	return mockBus
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
func createAndInsertDefaultTestApp(db *persistence.PostgresContext) (*coreTypes.Actor, error) {
	app, err := newTestApp()
	if err != nil {
		return nil, err
	}
	// TODO(andrew): Avoid the use of `log.Fatal(fmt.Sprintf`
	// TODO(andrew): Use `require.NoError` instead of `log.Fatal` in tests`
	addrBz, err := hex.DecodeString(app.Address)
	if err != nil {
		log.Fatalf("an error occurred converting address to bytes %s", app.Address)
	}
	pubKeyBz, err := hex.DecodeString(app.PublicKey)
	if err != nil {
		log.Fatalf("an error occurred converting pubKey to bytes %s", app.PublicKey)
	}
	outputBz, err := hex.DecodeString(app.Output)
	if err != nil {
		log.Fatalf("an error occurred converting output to bytes %s", app.Output)
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
