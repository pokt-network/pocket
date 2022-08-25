package test

import (
	"encoding/json"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	types2 "github.com/pokt-network/pocket/utility/types"
	"log"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
)

var (
	defaultTestingChainsEdited = []string{"0002"}
	defaultUnstaking           = int64(2017)
	defaultSendAmount          = big.NewInt(10000)
	defaultNonceString         = types2.BigIntToString(test_artifacts.DefaultAccountAmount)
	defaultSendAmountString    = types2.BigIntToString(defaultSendAmount)
	testSchema                 = "test_schema"
)
var testPersistenceMod modules.PersistenceModule

func NewTestingMempool(_ *testing.T) types2.Mempool {
	return types2.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource, dbUrl := test_artifacts.SetupPostgresDocker()
	testPersistenceMod = newTestPersistenceModule(dbUrl)
	m.Run()
	test_artifacts.CleanupPostgresDocker(m, pool, resource)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
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

// TODO_IN_THIS_COMMIT: Take in `t` or return an error
func newTestPersistenceModule(databaseUrl string) modules.PersistenceModule {
	cfg := modules.Config{
		Persistence: &test_artifacts.MockPersistenceConfig{
			PostgresUrl:    databaseUrl,
			NodeSchema:     testSchema,
			BlockStorePath: "",
		},
	}
	genesisState, _ := test_artifacts.NewGenesisState(5, 1, 1, 1)
	config, _ := json.Marshal(cfg.Persistence)
	genesis, _ := json.Marshal(genesisState.PersistenceGenesisState)
	persistenceMod, err := persistence.Create(config, genesis) // TODO (Drewsky) this is the last remaining cross module import and needs a fix...
	if err != nil {
		log.Fatalf("Error creating persistence module: %s", err)
	}
	persistenceMod.Start() // TODO: Check for error
	return persistenceMod
}
