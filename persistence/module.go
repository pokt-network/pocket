package persistence

import (
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"math"
	"math/big"
	"pocket/consensus/pkg/config"
	"pocket/shared/context"
	"pocket/shared/modules"

	// REALLY BAD. WE SHOULD NEVER DO THIS.
	"pocket/utility/utility/test"
	"pocket/utility/utility/types"
)

type persistenceModule struct {
	modules.PersistenceModule

	CommitDB     *memdb.DB
	Mempool      types.Mempool
	pocketBusMod modules.PocketBusModule
}

func Create(config *config.Config) (modules.PersistenceModule, error) {
	db := memdb.New(comparer.DefaultComparer, 888888888)

	return &persistenceModule{
		PersistenceModule: nil,
		CommitDB:          db,
		Mempool:           types.NewMempool(1000, 1000),
		pocketBusMod:      nil,
	}, nil

}

func (p *persistenceModule) Start(ctx *context.PocketContext) error {
	pCtx, err := p.NewContext(0)
	if err != nil {
		return err
	}
	genesis, _, _, _, _, err := test.NewMockGenesisState(5, 1, 1, 5)
	if err != nil {
		return err
	}
	c := pCtx.(*test.MockPersistenceContext)
	err = InitGenesis(c, genesis)
	if err != nil {
		return err
	}
	err = c.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (p *persistenceModule) Stop(*context.PocketContext) error {
	return nil
}

func (m *persistenceModule) SetPocketBusMod(pocketBus modules.PocketBusModule) {
	m.pocketBusMod = pocketBus
}

func (m *persistenceModule) GetPocketBusMod() modules.PocketBusModule {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (m *persistenceModule) NewContext(height int64) (modules.PersistenceContext, error) {
	newDB := test.NewMemDB()
	it := m.CommitDB.NewIterator(&util.Range{Start: test.HeightKey(height, nil), Limit: test.HeightKey(height+1, nil)})
	it.First()
	defer it.Release()
	for ; it.Valid(); it.Next() {
		err := newDB.Put(test.KeyFromHeightKey(it.Key()), it.Value())
		if err != nil {
			return nil, err
		}
	}
	context := &test.MockPersistenceContext{
		Height:     height,
		Parent:     m,
		SavePoints: make(map[string]int),
		DBs:        make([]*memdb.DB, 0),
	}
	context.SavePoints[hex.EncodeToString(test.FirstSavePointKey)] = types.ZeroInt
	context.DBs = append(context.DBs, newDB)
	return context, nil
}

func (m *persistenceModule) GetCommitDB() *memdb.DB {
	return m.CommitDB
}

func InitGenesis(u *test.MockPersistenceContext, state *test.GenesisState) error {
	if err := test.InsertPersistenceParams(u, state.Params); err != nil {
		return err
	}
	for _, account := range state.Accounts {
		if err := u.SetAccount(account.Address, account.Amount); err != nil {
			return err
		}
	}
	for _, p := range state.Pools {
		if err := u.InsertPool(p.Name, p.Account.Address, p.Account.Amount); err != nil {
			return err
		}
	}
	for _, validator := range state.Validators {
		err := u.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 2, validator.ServiceURL, validator.StakedTokens, 0, 0)
		if err != nil {
			return err
		}
	}
	for _, fisherman := range state.Fishermen {
		err := u.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 2, fisherman.ServiceURL, fisherman.StakedTokens, fisherman.Chains, 0, 0)
		if err != nil {
			return err
		}
	}
	for _, serviceNode := range state.ServiceNodes {
		err := u.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 2, serviceNode.ServiceURL, serviceNode.StakedTokens, serviceNode.Chains, 0, 0)
		if err != nil {
			return err
		}
	}
	for _, application := range state.Apps {
		maxRelays, err := CalculateAppRelays(u, 0, application.StakedTokens)
		if err != nil {
			return err
		}
		err = u.InsertApplication(application.Address, application.PublicKey, application.Output, false, 2, maxRelays, application.StakedTokens, application.Chains, 0, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO
func CalculateAppRelays(u *test.MockPersistenceContext, height int64, stakedTokens string) (string, error) {
	tokens, err := StringToBigInt(stakedTokens)
	if err != nil {
		return types.EmptyString, err
	}
	p, er := u.GetParams(height)
	if er != nil {
		return "", err
	}
	stakingAdjustment := p.GetAppStakingAdjustment()
	if err != nil {
		return types.EmptyString, err
	}
	baseRate := p.GetAppBaselineStakeRate()
	if err != nil {
		return types.EmptyString, err
	}
	// convert tokens to int64
	tokensFloat64 := big.NewFloat(float64(tokens.Int64()))
	// get the percentage of the baseline stake rate (can be over 100%)
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))
	// multiply the two
	baselineThroughput := basePercentage.Mul(basePercentage, tokensFloat64)
	// adjust for uPOKT
	baselineThroughput.Quo(baselineThroughput, big.NewFloat(1000000))
	// add staking adjustment (can be negative)
	adjusted := baselineThroughput.Add(baselineThroughput, big.NewFloat(float64(stakingAdjustment)))
	// truncate the integer
	result, _ := adjusted.Int(nil)
	// bounding Max Amount of relays to maxint64
	max := big.NewInt(math.MaxInt64)
	if i := result.Cmp(max); i < -1 {
		result = max
	}
	return BigIntToString(result), nil
}

func StringToBigInt(s string) (*big.Int, types.Error) {
	b := big.NewInt(0)
	i, ok := b.SetString(s, 10)
	if !ok {
		return nil, types.ErrStringToBigInt()
	}
	return i, nil
}

func BigIntToString(b *big.Int) string {
	return b.Text(10)
}
