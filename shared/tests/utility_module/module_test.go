package utility_module

import (
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	"math/big"
	"testing"

<<<<<<< HEAD
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility"
=======
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
>>>>>>> main
)

var (
	defaultTestingChainsEdited = []string{"0002"}
	defaultUnstaking           = int64(2017)
	defaultSendAmount          = big.NewInt(10000)
	defaultNonceString         = types.BigIntToString(test_artifacts.DefaultAccountAmount)
	defaultSendAmountString    = types.BigIntToString(defaultSendAmount)
)

func NewTestingMempool(_ *testing.T) types.Mempool {
	return types.NewMempool(1000000, 1000)
}

<<<<<<< HEAD
func TestMain(m *testing.M) {
	pool, resource := tests.SetupPostgresDocker()
=======
var testPersistenceMod modules.PersistenceModule

func TestMain(m *testing.M) {
	pool, resource, mod := tests.SetupPostgresDockerPersistenceMod()
	testPersistenceMod = mod
>>>>>>> main
	m.Run()
	tests.CleanupPostgresDocker(m, pool, resource)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
<<<<<<< HEAD
	mempool := NewTestingMempool(t)
	cfg := &genesis.Config{
		Base:      &genesis.BaseConfig{},
		Consensus: &genesis.ConsensusConfig{},
		Utility:   &genesis.UtilityConfig{},
		Persistence: &genesis.PersistenceConfig{
			PostgresUrl:    tests.DatabaseUrl,
			NodeSchema:     tests.SQL_Schema,
			BlockStorePath: "",
		},
		P2P:       &genesis.P2PConfig{},
		Telemetry: &genesis.TelemetryConfig{},
	}
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	var err error
	tests.PersistenceModule, err = persistence.Create(cfg, genesisState)
	require.NoError(t, err)
	require.NoError(t, tests.PersistenceModule.Start(), "start persistence mod")
	persistenceContext, err := tests.PersistenceModule.NewRWContext(height)
=======
	persistenceContext, err := testPersistenceMod.NewRWContext(height)
>>>>>>> main
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
