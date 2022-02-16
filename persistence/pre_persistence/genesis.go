package pre_persistence

import (
	"math"
	"math/big"
)

func InitGenesis(u *MockPersistenceContext, state *GenesisState) error {
	if err := InsertPersistenceParams(u, state.Params); err != nil {
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
func CalculateAppRelays(u *MockPersistenceContext, height int64, stakedTokens string) (string, error) {
	tokens, err := StringToBigInt(stakedTokens)
	if err != nil {
		return EmptyString, err
	}
	p, er := u.GetParams(height)
	if er != nil {
		return "", err
	}
	stakingAdjustment := p.GetAppStakingAdjustment()
	if err != nil {
		return EmptyString, err
	}
	baseRate := p.GetAppBaselineStakeRate()
	if err != nil {
		return EmptyString, err
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
