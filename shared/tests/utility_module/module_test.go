package utility_module

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/tests"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility"
)

var (
	defaultTestingChains       = []string{"0001"}
	defaultTestingChainsEdited = []string{"0002"}
	defaultAmount              = big.NewInt(1000000000000000)
	defaultSendAmount          = big.NewInt(10000)
	defaultAmountString        = types.BigIntToString(defaultAmount)
	defaultNonceString         = types.BigIntToString(defaultAmount)
	defaultSendAmountString    = types.BigIntToString(defaultSendAmount)
)

func NewTestingMempool(_ *testing.T) types.Mempool {
	return types.NewMempool(1000000, 1000)
}

func TestMain(m *testing.M) {
	pool, resource := tests.SetupPostgresDocker()
	m.Run()
	tests.CleanupPostgresDocker(m, pool, resource)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	mempool := NewTestingMempool(t)
	cfg := &config.Config{
		RootDir: "",
		GenesisSource: &genesis.GenesisSource{
			Source: &genesis.GenesisSource_Config{
				Config: genesisConfig(),
			},
		},
		Persistence: &config.PersistenceConfig{
			PostgresUrl: tests.DatabaseUrl,
			NodeSchema:  tests.SQLSchema,
		},
	}
	err := cfg.HydrateGenesisState()
	require.NoError(t, err)

	tests.PersistenceModule, err = persistence.Create(cfg)
	require.NoError(t, err)
	require.NoError(t, tests.PersistenceModule.Start(), "start persistence mod")

	persistenceContext, err := tests.PersistenceModule.NewRWContext(height)
	require.NoError(t, err)

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

func genesisConfig() *genesis.GenesisConfig {
	config := &genesis.GenesisConfig{
		NumValidators:   5,
		NumApplications: 1,
		NumFisherman:    1,
		NumServicers:    1,
	}
	return config
}
