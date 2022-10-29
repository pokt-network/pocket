package test

import (
	"encoding/hex"
	"math/big"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
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
	defaultNonceString         = utilTypes.BigIntToString(test_artifacts.DefaultAccountAmount)
	defaultSendAmountString    = utilTypes.BigIntToString(defaultSendAmount)
	testSchema                 = "test_schema"
	testMessageSendType        = "MessageSend"
)

var persistenceDbUrl string
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
	persistenceDbUrl = dbUrl
	exitCode := m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	// IMPROVE: Avoid creating a new persistence module with every test
	testPersistenceMod := newTestPersistenceModule(t, persistenceDbUrl)

	persistenceContext, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testPersistenceMod.ReleaseWriteContext())
		require.NoError(t, testPersistenceMod.HandleDebugMessage(&debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE,
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

func newTestPersistenceModule(t *testing.T, databaseUrl string) modules.PersistenceModule {
	ctrl := gomock.NewController(t)

	mockPersistenceConfig := mock_modules.NewMockPersistenceConfig(ctrl)
	mockPersistenceConfig.EXPECT().GetPostgresUrl().Return(databaseUrl).AnyTimes()
	mockPersistenceConfig.EXPECT().GetNodeSchema().Return(testSchema).AnyTimes()
	mockPersistenceConfig.EXPECT().GetBlockStorePath().Return("").AnyTimes()
	mockPersistenceConfig.EXPECT().GetTxIndexerPath().Return("").AnyTimes()

	mockRuntimeConfig := mock_modules.NewMockConfig(ctrl)
	mockRuntimeConfig.EXPECT().GetPersistenceConfig().Return(mockPersistenceConfig).AnyTimes()

	mockRuntimeMgr := mock_modules.NewMockRuntimeMgr(ctrl)
	mockRuntimeMgr.EXPECT().GetConfig().Return(mockRuntimeConfig).AnyTimes()

	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	mockRuntimeMgr.EXPECT().GetGenesis().Return(genesisState).AnyTimes()

	persistenceMod, err := persistence.Create(mockRuntimeMgr)
	require.NoError(t, err)

	err = persistenceMod.Start()
	require.NoError(t, err)

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
