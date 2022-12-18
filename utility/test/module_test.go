package test

import (
	"encoding/hex"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/utility"
	utilTypes "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

const (
	testingValidatorCount   = 5
	testingServiceNodeCount = 1
	testingApplicationCount = 1
	testingFishermenCount   = 1
)

var (
	defaultTestingChainsEdited = []string{"0002"}

	defaultUnstaking   = int64(2017)
	defaultNonceString = utilTypes.BigIntToString(defaults.DefaultAccountAmount)

	testNonce           = "defaultNonceString"
	testSchema          = "test_schema"
	testMessageSendType = "MessageSend"
)

var testPersistenceMod modules.PersistenceModule // initialized in TestMain
var testUtilityMod modules.UtilityModule         // initialized in TestMain

var actorTypes = []utilTypes.ActorType{
	utilTypes.ActorType_App,
	utilTypes.ActorType_ServiceNode,
	utilTypes.ActorType_Fisherman,
	utilTypes.ActorType_Validator,
}

func NewTestingMempool(_ *testing.T) utilTypes.Mempool {
	return utilTypes.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	runtimeCfg := newTestRuntimeConfig(dbUrl)

	testUtilityMod = newTestUtilityModule(runtimeCfg)
	testPersistenceMod = newTestPersistenceModule(runtimeCfg)

	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	persistenceContext, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	// TODO(#388): Expose a `GetMempool` function in `utility_module` so we can remove this reflection.
	mempool := reflect.ValueOf(testUtilityMod).Elem().FieldByName("Mempool").Interface().(utilTypes.Mempool)

	// TECHDEBT: Move the internal of cleanup into a separate function and call this in the
	// beginning of every test. This (the current implementation) is an issue because if we call
	// `NewTestingUtilityContext` more than once in a single test, we create unnecessary calls to clean.
	t.Cleanup(func() {
		require.NoError(t, testPersistenceMod.ReleaseWriteContext())
		require.NoError(t, testPersistenceMod.HandleDebugMessage(&messaging.DebugMessage{
			Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
			Message: nil,
		}))
		mempool.Clear()
	})

	return utility.UtilityContext{
		Height:  height,
		Mempool: mempool,
		Context: &utility.Context{
			PersistenceRWContext: persistenceContext,
			SavePointsM:          make(map[string]struct{}),
			SavePoints:           make([][]byte, 0),
		},
	}
}

func newTestRuntimeConfig(databaseUrl string) *runtime.Manager {
	cfg := runtime.NewConfig(
		&runtime.BaseConfig{},
		runtime.WithPersistenceConfig(&types.PersistenceConfig{
			PostgresUrl:       databaseUrl,
			NodeSchema:        testSchema,
			BlockStorePath:    "",
			TxIndexerPath:     "",
			TreesStoreDir:     "",
			MaxConns:          4,
			MinConns:          0,
			MaxConnLifetime:   1,
			MaxConnIdleTime:   30,
			HealthCheckPeriod: 5,
		}),
		runtime.WithUtilityConfig(&utilTypes.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
		}))
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	runtimeCfg := runtime.NewManager(cfg, genesisState)
	return runtimeCfg
}

func newTestUtilityModule(runtimeCfg *runtime.Manager) modules.UtilityModule {
	utilityMod, err := utility.Create(runtimeCfg)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return utilityMod.(modules.UtilityModule)
}

// IMPROVE: Not part of `TestMain` because a mock requires `testing.T` to be initialized.
// We are trying to only initialize one `testPersistenceModule` in all the tests, so when the
// utility module tests are no longer dependant on the persistence module explicitly, this
// can be improved.
func mockBusInTestModules(t *testing.T) {
	ctrl := gomock.NewController(t)

	busMock := modulesMock.NewMockBus(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(testPersistenceMod).AnyTimes()
	busMock.EXPECT().GetUtilityModule().Return(testUtilityMod).AnyTimes()

	testPersistenceMod.SetBus(busMock)
	testUtilityMod.SetBus(busMock)

	t.Cleanup(func() {
		testPersistenceMod.SetBus(nil)
		testUtilityMod.SetBus(nil)
	})
}

// TODO(#290): Mock the persistence module so the utility module is not dependant on it.
func newTestPersistenceModule(runtimeCfg *runtime.Manager) modules.PersistenceModule {
	persistenceMod, err := persistence.Create(runtimeCfg)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod.(modules.PersistenceModule)
}

func requireValidTestingTxResults(t *testing.T, tx *utilTypes.Transaction, txResults []modules.TxResult) {
	for _, txResult := range txResults {
		msg, err := tx.GetMessage()
		sendMsg, ok := msg.(*utilTypes.MessageSend)
		require.True(t, ok)
		require.NoError(t, err)
		require.Equal(t, int32(0), txResult.GetResultCode())
		require.Equal(t, "", txResult.GetError())
		require.Equal(t, testMessageSendType, txResult.GetMessageType())
		require.Equal(t, int32(0), txResult.GetIndex())
		require.Equal(t, int64(0), txResult.GetHeight())
		require.Equal(t, hex.EncodeToString(sendMsg.ToAddress), txResult.GetRecipientAddr())
		require.Equal(t, hex.EncodeToString(sendMsg.FromAddress), txResult.GetSignerAddr())
	}
}
