package pre_persistence

import (
	"log"
	"math"
	"math/big"

	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

func Create(cfg *config.Config) (modules.PersistenceModule, error) {
	db := memdb.New(comparer.DefaultComparer, cfg.PrePersistence.Capacity)
	return NewPrePersistenceModule(db, types.NewMempool(cfg.PrePersistence.MempoolMaxBytes, cfg.PrePersistence.MempoolMaxTxs), cfg), nil
}

func (p *PrePersistenceModule) Start() error {
	pCtx, err := p.NewContext(0)
	if err != nil {
		return err
	}
	c := pCtx.(*PrePersistenceContext)
	// TODO(team): Load saved state from disk instead of genesis
	err = InitGenesis(c, p.Cfg.GenesisSource.GetState())
	if err != nil {
		return err
	}
	err = c.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (p *PrePersistenceModule) Stop() error {
	return nil
}

func (m *PrePersistenceModule) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *PrePersistenceModule) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func InitGenesis(u *PrePersistenceContext, state *typesGenesis.GenesisState) error {
	if err := InsertPersistenceParams(u, state.Params); err != nil {
		return err
	}
	for _, account := range state.Accounts {
		if err := u.SetAccountAmount(account.Address, account.Amount); err != nil {
			return err
		}
	}
	for _, p := range state.Pools {
		if err := u.InsertPool(p.Name, p.Account.Address, p.Account.Amount); err != nil {
			return err
		}
	}
	for _, validator := range state.Validators {
		err := u.InsertValidator(validator.Address, validator.PublicKey, validator.Output, false, 2, validator.ServiceUrl, validator.StakedTokens, 0, 0)
		if err != nil {
			return err
		}
	}
	for _, fisherman := range state.Fishermen {
		err := u.InsertFisherman(fisherman.Address, fisherman.PublicKey, fisherman.Output, false, 2, fisherman.ServiceUrl, fisherman.StakedTokens, fisherman.Chains, 0, 0)
		if err != nil {
			return err
		}
	}
	for _, serviceNode := range state.ServiceNodes {
		err := u.InsertServiceNode(serviceNode.Address, serviceNode.PublicKey, serviceNode.Output, false, 2, serviceNode.ServiceUrl, serviceNode.StakedTokens, serviceNode.Chains, 0, 0)
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

// TODO(andrew): this is a state operation that really shouldn't live here, rather the utility module... but is needed for genesis creation
func CalculateAppRelays(u *PrePersistenceContext, height int64, stakedTokens string) (string, error) {
	tokens, err := types.StringToBigInt(stakedTokens)
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
	return types.BigIntToString(result), nil
}
