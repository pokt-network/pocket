package utility_module

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

var (
	defaultTestingChains          = []string{"0001"}
	defaultTestingChainsEdited    = []string{"0002"}
	defaultServiceUrl             = "https://foo.bar"
	defaultServiceUrlEdited       = "https://bar.foo"
	defaultServiceNodesPerSession = 24
	zeroAmount                    = big.NewInt(0)
	zeroAmountString              = types.BigIntToString(zeroAmount)
	defaultAmount                 = big.NewInt(1000000000000000)
	defaultSendAmount             = big.NewInt(10000)
	defaultAmountString           = types.BigIntToString(defaultAmount)
	defaultNonceString            = types.BigIntToString(defaultAmount)
	defaultSendAmountString       = types.BigIntToString(defaultSendAmount)
)

func NewTestingMempool(_ *testing.T) types.Mempool {
	return types.NewMempool(1000000, 1000)
}

func NewTestingUtilityContext(t *testing.T, height int64) utility.UtilityContext {
	mempool := NewTestingMempool(t)
	persistenceModule := pre_persistence.NewPrePersistenceModule(memdb.New(comparer.DefaultComparer, 10000000), mempool, &config.Config{IsTesting: true})
	if err := persistenceModule.Start(); err != nil {
		t.Fatal(err)
	}
	persistenceContext, err := persistenceModule.NewContext(height)
	if err != nil {
		t.Fatal(err)
	}
	return utility.UtilityContext{
		LatestHeight: height,
		Mempool:      mempool,
		Context: &utility.Context{
			PersistenceContext: persistenceContext,
			SavePointsM:        make(map[string]struct{}),
			SavePoints:         make([][]byte, 0),
		},
	}
}
