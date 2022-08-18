package utility_module

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
)

var (
	defaultTestingChainsEdited = []string{"0002"}
	defaultUnstaking           = int64(2017)
	defaultSendAmount          = big.NewInt(10000)
	defaultNonceString         = types.BigIntToString(test_artifacts.DefaultAccountAmount)
	defaultSendAmountString    = types.BigIntToString(defaultSendAmount)
	testSchema                 = "test_schema"
)
var databaseUrl string // initialized in TestMain

func NewTestingMempool(_ *testing.T) types.Mempool {
	return types.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := tests.SetupPostgresDocker()
	databaseUrl = dbUrl
	m.Run()
	tests.CleanupPostgresDocker(m, pool, resource)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	mempool := NewTestingMempool(t)
	cfg := &genesis.Config{
		Base:      &genesis.BaseConfig{},
		Consensus: &genesis.ConsensusConfig{},
		Utility:   &genesis.UtilityConfig{},
		Persistence: &genesis.PersistenceConfig{
			PostgresUrl:    databaseUrl,
			NodeSchema:     testSchema,
			BlockStorePath: "",
		},
		P2P:       &genesis.P2PConfig{},
		Telemetry: &genesis.TelemetryConfig{},
	}
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	var err error
	persistenceMod, err := persistence.Create(cfg, genesisState)
	require.NoError(t, err)
	require.NoError(t, persistenceMod.Start(), "start persistence mod")
	persistenceContext, err := persistenceMod.NewRWContext(height)
	require.NoError(t, err)

	mempool := NewTestingMempool(t)
	return utility.UtilityContext{
		LatestHeight: height,
		Mempool:      mempool,
		Context: &utility.Context{
			PersistenceRWContext: persistenceContext,
			SavePointsM:          make(map[string]struct{}),
			SavePoints:           make([][]byte, 0),
		},
	}
}
