package test

import (
	"encoding/hex"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
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
	defaultUnstaking           = int64(2017)
	defaultSendAmount          = big.NewInt(10000)
	defaultNonceString         = utilTypes.BigIntToString(defaults.DefaultAccountAmount)
	defaultSendAmountString    = utilTypes.BigIntToString(defaultSendAmount)
	testSchema                 = "test_schema"
	testMessageSendType        = "MessageSend"
)

var testPersistenceMod modules.PersistenceModule // initialized in TestMain
var persistenceDbUrl string
var actorTypes = []utilTypes.ActorType{
	// utilTypes.ActorType_App,
	// utilTypes.ActorType_ServiceNode,
	utilTypes.ActorType_Fisherman,
	utilTypes.ActorType_Validator,
}

func NewTestingMempool(_ *testing.T) utilTypes.Mempool {
	return utilTypes.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(dbUrl)
	// persistenceDbUrl = dbUrl
	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	// IMPROVE: Avoid creating a new persistence module with every test
	// testPersistenceMod := newTestPersistenceModule(t, persistenceDbUrl)

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
	})

	return utility.UtilityContext{
		LatestHeight: height,
		Mempool:      NewTestingMempool(t),
		Context: &utility.Context{
			PersistenceRWContext: persistenceContext,
			SavePointsM:          make(map[string]struct{}),
			SavePoints:           make([][]byte, 0),
		},
	}
}

// TODO(olshansky): Take in `t testing.T` as a parameter and error if there's an issue
func newTestPersistenceModule(databaseUrl string) modules.PersistenceModule {
	// HACK: See `runtime/test_artifacts/generator.go` for why we're doing this to get deterministic key generation.
	// os.Setenv(test_artifacts.PrivateKeySeedEnv, "42")
	// defer os.Unsetenv(test_artifacts.PrivateKeySeedEnv)

	cfg := runtime.NewConfig(&runtime.BaseConfig{}, runtime.WithPersistenceConfig(&types.PersistenceConfig{
		PostgresUrl:    databaseUrl,
		NodeSchema:     testSchema,
		BlockStorePath: "",
		TxIndexerPath:  "",
		TreesStoreDir:  "",
	}))
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	runtimeCfg := runtime.NewManager(cfg, genesisState)

	persistenceMod, err := persistence.Create(runtimeCfg)
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	return persistenceMod.(modules.PersistenceModule)
}

// func newTestPersistenceModule(t *testing.T, databaseUrl string) modules.PersistenceModule {
// 	ctrl := gomock.NewController(t)

// 	mockPersistenceConfig := mock_modules.NewMockPersistenceConfig(ctrl)
// 	mockPersistenceConfig.EXPECT().GetPostgresUrl().Return(databaseUrl).AnyTimes()
// 	mockPersistenceConfig.EXPECT().GetNodeSchema().Return(testSchema).AnyTimes()
// 	mockPersistenceConfig.EXPECT().GetBlockStorePath().Return("").AnyTimes()
// 	mockPersistenceConfig.EXPECT().GetTxIndexerPath().Return("").AnyTimes()
// 	mockPersistenceConfig.EXPECT().GetTreesStoreDir().Return("").AnyTimes()

// 	mockRuntimeConfig := mock_modules.NewMockConfig(ctrl)
// 	mockRuntimeConfig.EXPECT().GetPersistenceConfig().Return(mockPersistenceConfig).AnyTimes()

// 	mockRuntimeMgr := mock_modules.NewMockRuntimeMgr(ctrl)
// 	mockRuntimeMgr.EXPECT().GetConfig().Return(mockRuntimeConfig).AnyTimes()

// 	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
// 	mockRuntimeMgr.EXPECT().GetGenesis().Return(genesisState).AnyTimes()

// 	persistenceMod, err := persistence.Create(mockRuntimeMgr)
// 	require.NoError(t, err)

// 	err = persistenceMod.Start()
// 	require.NoError(t, err)

// 	return persistenceMod.(modules.PersistenceModule)
// }

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
