package test

import (
	"math/big"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
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
	testPersistenceMod := newTestPersistenceModule(t, persistenceDbUrl)

	persistenceContext, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	t.Cleanup(func() {
		persistenceContext.ResetContext()
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

	mockPersistenceConfig := mockModules.NewMockPersistenceConfig(ctrl)
	mockPersistenceConfig.EXPECT().GetPostgresUrl().Return(databaseUrl).AnyTimes()
	mockPersistenceConfig.EXPECT().GetNodeSchema().Return(testSchema).AnyTimes()
	mockPersistenceConfig.EXPECT().GetBlockStorePath().Return("").AnyTimes()

	mockRuntimeConfig := mockModules.NewMockConfig(ctrl)
	mockRuntimeConfig.EXPECT().GetPersistenceConfig().Return(mockPersistenceConfig).AnyTimes()

	mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
	mockRuntimeMgr.EXPECT().GetConfig().Return(mockRuntimeConfig).AnyTimes()

	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	mockRuntimeMgr.EXPECT().GetGenesis().Return(genesisState).AnyTimes()

	persistenceMod, err := persistence.Create(mockRuntimeMgr)
	require.NoError(t, err)

	err = persistenceMod.Start()
	require.NoError(t, err)

	return persistenceMod.(modules.PersistenceModule)
}
