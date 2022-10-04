package test

import (
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	utilTypes "github.com/pokt-network/pocket/utility/types"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
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

func NewTestingMempool(_ *testing.T) utilTypes.Mempool {
	return utilTypes.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	persistenceDbUrl = dbUrl
	m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	testPersistenceMod := newTestPersistenceModule(t, persistenceDbUrl)

	persistenceContext, err := testPersistenceMod.NewRWContext(height)
	require.NoError(t, err)

	t.Cleanup(func() {
		testPersistenceMod.ResetContext()
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
