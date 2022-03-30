package pre_persistence

// TODO(team): Consolidate this `gensis.go` with `shared/genesis.go`

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/crypto"
)

var ( // TODO these are needed placeholders to pass validation checks. Until we have a real genesis implementation & testing environment, this will suffice
	defaultChains         = []string{"0001"}
	defaultServiceUrl     = "https://foo.bar"
	defaultStakeBig       = big.NewInt(1000000000000000)
	defaultStake          = BigIntToString(defaultStakeBig)
	defaultAccountbalance = defaultStake
	defaultStakeStatus    = int32(2)
)

// NewGenesisState IMPORTANT NOTE: Not using numOfValidators param, as Validators are now read from the test_state json file
func NewGenesisState(cfg *config.Config, numOfValidators, numOfApplications, numOfFisherman, numOfServiceNodes int) (state *GenesisState, validatorKeys, appKeys, serviceNodeKeys, fishKeys []crypto.PrivateKey, err error) {
	// create the genesis state object
	state = &GenesisState{}
	// use the `integration test state` to populate parts of the genesis state
	testingState := GetTestState()
	// populate genesis object with the 'test state' validator map
	vm := testingState.ValidatorMap
	// generate `mocked` keys for each actor
	// specifically, use the test state json file to create the validator keys
	validatorKeys = make([]crypto.PrivateKey, len(vm))
	// use the number param to create the rest
	appKeys = make([]crypto.PrivateKey, numOfApplications)
	fishKeys = make([]crypto.PrivateKey, numOfFisherman)
	serviceNodeKeys = make([]crypto.PrivateKey, numOfServiceNodes)
	// create state objects for each key type
	for i := range validatorKeys {
		var pk crypto.PrivateKey
		n := vm[NodeId(i+1)] // TODO will have to fix conflict when NodeId is deprecated
		pk, _ = crypto.NewPrivateKey(n.PrivateKey)
		v := &Validator{
			Status:       2,
			ServiceUrl:   defaultServiceUrl,
			StakedTokens: defaultStake,
		}
		v.Address = pk.Address()
		v.PublicKey = pk.PublicKey().Bytes()
		v.Output = v.Address
		state.Validators = append(state.Validators, v)
		state.Accounts = append(state.Accounts, &Account{
			Address: v.Address,
			Amount:  defaultAccountbalance,
		})
		validatorKeys[i] = pk
	}
	for i := range appKeys {
		pk, _ := crypto.GeneratePrivateKey()
		app := &App{
			Status:       defaultStakeStatus,
			Chains:       defaultChains,
			StakedTokens: defaultStake,
		}
		app.Address = pk.Address()
		app.PublicKey = pk.PublicKey().Bytes()
		app.Output = app.Address
		state.Apps = append(state.Apps, app)
		state.Accounts = append(state.Accounts, &Account{
			Address: app.Address,
			Amount:  defaultAccountbalance,
		})
		appKeys[i] = pk
	}
	for i := range serviceNodeKeys {
		pk, _ := crypto.GeneratePrivateKey()
		sn := &ServiceNode{
			Status:       defaultStakeStatus,
			ServiceUrl:   defaultServiceUrl,
			Chains:       defaultChains,
			StakedTokens: defaultStake,
		}
		sn.Address = pk.Address()
		sn.PublicKey = pk.PublicKey().Bytes()
		sn.Output = sn.Address
		state.ServiceNodes = append(state.ServiceNodes, sn)
		state.Accounts = append(state.Accounts, &Account{
			Address: sn.Address,
			Amount:  defaultAccountbalance,
		})
		serviceNodeKeys[i] = pk
	}
	for i := range fishKeys {
		pk, _ := crypto.GeneratePrivateKey()
		fish := &Fisherman{
			Status:       defaultStakeStatus,
			Chains:       defaultChains,
			ServiceUrl:   defaultServiceUrl,
			StakedTokens: defaultStake,
		}
		fish.Address = pk.Address()
		fish.PublicKey = pk.PublicKey().Bytes()
		fish.Output = fish.Address
		state.Fishermen = append(state.Fishermen, fish)
		state.Accounts = append(state.Accounts, &Account{
			Address: fish.Address,
			Amount:  defaultAccountbalance,
		})
		fishKeys[i] = pk
	}
	// populate the state with default parameters
	state.Params = DefaultParams()
	// create appropriate 'stake' pools for each actor type
	valStakePool, err := NewPool(ValidatorStakePoolName, &Account{
		Address: DefaultValidatorStakePool.Address(),
		Amount:  BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	appStakePool, err := NewPool(AppStakePoolName, &Account{
		Address: DefaultAppStakePool.Address(),
		Amount:  BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	fishStakePool, err := NewPool(FishermanStakePoolName, &Account{
		Address: DefaultFishermanStakePool.Address(),
		Amount:  BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	serNodeStakePool, err := NewPool(ServiceNodeStakePoolName, &Account{
		Address: DefaultServiceNodeStakePool.Address(),
		Amount:  BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	// create a pool for collected fees (helps with rewards)
	fee, err := NewPool(FeePoolName, &Account{
		Address: DefaultFeeCollector.Address(),
		Amount:  BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	// create a pool for the dao treasury
	dao, err := NewPool(DAOPoolName, &Account{
		Address: DefaultDAOPool.Address(),
		Amount:  BigIntToString(big.NewInt(0)),
	})
	if err != nil {
		return
	}
	// create an account for the DAO / Param owner
	pOwnerAddress := DefaultParamsOwner.Address()
	state.Accounts = append(state.Accounts, &Account{
		Address: pOwnerAddress,
		Amount:  defaultAccountbalance,
	})
	// populate the state pools with the previously created
	state.Pools = append(state.Pools, dao)
	state.Pools = append(state.Pools, fee)
	state.Pools = append(state.Pools, serNodeStakePool)
	state.Pools = append(state.Pools, fishStakePool)
	state.Pools = append(state.Pools, appStakePool)
	state.Pools = append(state.Pools, valStakePool)
	return
}

func InitGenesis(u *PrePersistenceContext, state *GenesisState) error {
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

// TODO this is a state operation that really shouldn't live here, rather the utility module... but is needed for genesis creation
func CalculateAppRelays(u *PrePersistenceContext, height int64, stakedTokens string) (string, error) {
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
