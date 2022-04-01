package genesis

// TODO(team): Consolidate this with `shared/genesis.go`

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"github.com/pokt-network/pocket/shared/types"
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
)

const ( // Names for each 'pool' (specialized accounts)
	ServiceNodeStakePoolName = "SERVICE_NODE_STAKE_POOL"
	AppStakePoolName         = "APP_STAKE_POOL"
	ValidatorStakePoolName   = "VALIDATOR_STAKE_POOL"
	FishermanStakePoolName   = "FISHERMAN_STAKE_POOL"
	DAOPoolName              = "DAO_POOL"
	FeePoolName              = "FEE_POOL"
)

var (
	// NOTE: this is for fun illustration purposes... The addresses begin with DA0, DA0, and FEE :)
	// Of course, in a production network the params / owners must be set in the genesis file
	DefaultParamsOwner, _          = crypto.NewPrivateKey("ff538589deb7f28bbce1ba68b37d2efc0eaa03204b36513cf88422a875559e38d6cbe0430ddd85a5e48e0c99ef3dea47bf0d1a83c6e6ad1640f72201dc8a0120")
	DefaultDAOPool, _              = crypto.NewPrivateKey("b1dfb25a67dadf9cdd39927b86166149727649af3a3143e66e558652f8031f3faacaa24a69bcf2819ed97ab5ed8d1e490041e5c7ef9e1eddba8b5678f997ae58")
	DefaultFeeCollector, _         = crypto.NewPrivateKey("bdc02826b5da77b90a5d1550443b3f007725cc654c10002aa01e65a131f3464b826f8e7911fa89b4bd6659c3175114d714c60bac63acc63817c0d3a4ed2fdab8")
	DefaultFishermanStakePool, _   = crypto.NewPrivateKey("f3dd5c8ccd9a7c8d0afd36424c6fbe8ead55315086ef3d0d03ce8c7357e5e306733a711adb6fc8fbef6a3e2ac2db7842433053a23c751d19573ab85b52316f67")
	DefaultServiceNodeStakePool, _ = crypto.NewPrivateKey("b4e4426ed014d5ee89949e6f60c406c328e4fce466cd25f4697a41046b34313097a8cc38033822da010422851062ae6b21b8e29d4c34193b7d8fa0f37b6593b6")
	DefaultValidatorStakePool, _   = crypto.NewPrivateKey("e0b8b7cdb33f11a8d70eb05070e53b02fe74f4499aed7b159bd2dd256e356d67664b5b682e40ee218e5feea05c2a1bb595ec15f3850c92b571cdf950b4d9ba23")
	DefaultAppStakePool, _         = crypto.NewPrivateKey("429627bac8dc322f0aeeb2b8f25b329899b7ebb9605d603b5fb74557b13357e50834e9575c19d9d7d664ec460a98abb2435ece93440eb482c87d5b7259a8d271")
)

var ( // TODO these are needed placeholders to pass validation checks. Until we have a real genesis implementation & testing environment, this will suffice
	DefaultChains         = []string{"0001"}
	DefaultServiceUrl     = "https://foo.bar"
	DefaultStakeBig       = big.NewInt(1000000000000000)
	DefaultStake          = types.BigIntToString(DefaultStakeBig)
	DefaultAccountBalance = DefaultStake
	DefaultStakeStatus    = int32(2)
)

// TODO(team): NewGenesisStateConfigs is ONLY used for development purposes and disregards the
// other configs in the genesis file if specified. it is used to seed and configure data in
// `NewGenesisState`.
type NewGenesisStateConfigs struct {
	NumValidators    uint16 `json:"num_validators"`
	NumAppplications uint16 `json:"num_applications"`
	NumFisherman     uint16 `json:"num_fisherman"`
	NumServicers     uint16 `json:"num_servicers"`

	SeedStart          uint32 `json:"keys_seed_start"`
	ValidatorUrlFormat string `json:"validator_url_format"`
}

// NewGenesisState IMPORTANT NOTE: Not using numOfValidators param, as Validators are now read from the test_state json file
func NewGenesisState(genesisConfig *NewGenesisStateConfigs) (state *GenesisState, validatorKeys, appKeys, serviceNodeKeys, fishKeys []crypto.PrivateKey, err error) {
	// create the genesis state object
	state = &GenesisState{}
	validatorKeys = make([]crypto.PrivateKey, genesisConfig.NumValidators)
	appKeys = make([]crypto.PrivateKey, genesisConfig.NumAppplications)
	fishKeys = make([]crypto.PrivateKey, genesisConfig.NumFisherman)
	serviceNodeKeys = make([]crypto.PrivateKey, genesisConfig.NumServicers)
	seedNum := genesisConfig.SeedStart
	seed := make([]byte, ed25519.PrivateKeySize)
	// create state objects for each key type

	for i := range validatorKeys {
		seedNum++
		binary.LittleEndian.PutUint32(seed, seedNum)
		pk, err := crypto.NewPrivateKeyFromSeed(seed)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		var serviceUrl string
		if len(genesisConfig.ValidatorUrlFormat) > 0 {
			serviceUrl = fmt.Sprintf(genesisConfig.ValidatorUrlFormat, i+1)
		} else {
			serviceUrl = DefaultServiceUrl
		}
		v := &Validator{
			Status:       2, // TODO: What does this status mean?
			ServiceUrl:   serviceUrl,
			StakedTokens: DefaultStake,
		}
		v.Address = pk.Address()
		v.PublicKey = pk.PublicKey().Bytes()
		v.Output = v.Address
		state.Validators = append(state.Validators, v)
		state.Accounts = append(state.Accounts, &Account{
			Address: v.Address,
			Amount:  DefaultAccountBalance,
		})
		validatorKeys[i] = pk
	}
	for i := range appKeys {
		seedNum++
		binary.LittleEndian.PutUint32(seed, seedNum)
		pk, err := crypto.NewPrivateKeyFromSeed(seed)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		app := &App{
			Status:       DefaultStakeStatus,
			Chains:       DefaultChains,
			StakedTokens: DefaultStake,
		}
		app.Address = pk.Address()
		app.PublicKey = pk.PublicKey().Bytes()
		app.Output = app.Address
		state.Apps = append(state.Apps, app)
		state.Accounts = append(state.Accounts, &Account{
			Address: app.Address,
			Amount:  DefaultAccountBalance,
		})
		appKeys[i] = pk
	}
	for i := range serviceNodeKeys {
		seedNum++
		binary.LittleEndian.PutUint32(seed, seedNum)
		pk, err := crypto.NewPrivateKeyFromSeed(seed)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		sn := &ServiceNode{
			Status:       DefaultStakeStatus,
			ServiceUrl:   DefaultServiceUrl,
			Chains:       DefaultChains,
			StakedTokens: DefaultStake,
		}
		sn.Address = pk.Address()
		sn.PublicKey = pk.PublicKey().Bytes()
		sn.Output = sn.Address
		state.ServiceNodes = append(state.ServiceNodes, sn)
		state.Accounts = append(state.Accounts, &Account{
			Address: sn.Address,
			Amount:  DefaultAccountBalance,
		})
		serviceNodeKeys[i] = pk
	}
	for i := range fishKeys {
		seedNum++
		binary.LittleEndian.PutUint32(seed, seedNum)
		pk, err := crypto.NewPrivateKeyFromSeed(seed)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		fish := &Fisherman{
			Status:       DefaultStakeStatus,
			Chains:       DefaultChains,
			ServiceUrl:   DefaultServiceUrl,
			StakedTokens: DefaultStake,
		}
		fish.Address = pk.Address()
		fish.PublicKey = pk.PublicKey().Bytes()
		fish.Output = fish.Address
		state.Fishermen = append(state.Fishermen, fish)
		state.Accounts = append(state.Accounts, &Account{
			Address: fish.Address,
			Amount:  DefaultAccountBalance,
		})
		fishKeys[i] = pk
	}
	// populate the state with default parameters
	state.Params = DefaultParams()
	// create appropriate 'stake' pools for each actor type
	valStakePool, err := NewPool(ValidatorStakePoolName, &Account{
		Address: DefaultValidatorStakePool.Address(),
		Amount:  types.BigIntToString(&big.Int{}),
	})
	if err != nil {
		return
	}
	appStakePool, err := NewPool(AppStakePoolName, &Account{
		Address: DefaultAppStakePool.Address(),
		Amount:  types.BigIntToString(&big.Int{}),
	})
	if err != nil {
		return
	}
	fishStakePool, err := NewPool(FishermanStakePoolName, &Account{
		Address: DefaultFishermanStakePool.Address(),
		Amount:  types.BigIntToString(&big.Int{}),
	})
	if err != nil {
		return
	}
	serNodeStakePool, err := NewPool(ServiceNodeStakePoolName, &Account{
		Address: DefaultServiceNodeStakePool.Address(),
		Amount:  types.BigIntToString(&big.Int{}),
	})
	if err != nil {
		return
	}
	// create a pool for collected fees (helps with rewards)
	fee, err := NewPool(FeePoolName, &Account{
		Address: DefaultFeeCollector.Address(),
		Amount:  types.BigIntToString(&big.Int{}),
	})
	if err != nil {
		return
	}
	// create a pool for the dao treasury
	dao, err := NewPool(DAOPoolName, &Account{
		Address: DefaultDAOPool.Address(),
		Amount:  types.BigIntToString(&big.Int{}),
	})
	if err != nil {
		return
	}
	// create an account for the DAO / Param owner
	pOwnerAddress := DefaultParamsOwner.Address()
	state.Accounts = append(state.Accounts, &Account{
		Address: pOwnerAddress,
		Amount:  DefaultAccountBalance,
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

func DefaultParams() *Params {
	return &Params{
		BlocksPerSession:                         4,
		AppMinimumStake:                          types.BigIntToString(big.NewInt(15000000000)),
		AppMaxChains:                             15,
		AppBaselineStakeRate:                     100,
		AppStakingAdjustment:                     0,
		AppUnstakingBlocks:                       2016,
		AppMinimumPauseBlocks:                    4,
		AppMaxPauseBlocks:                        672,
		ServiceNodeMinimumStake:                  types.BigIntToString(big.NewInt(15000000000)),
		ServiceNodeMaxChains:                     15,
		ServiceNodeUnstakingBlocks:               2016,
		ServiceNodeMinimumPauseBlocks:            4,
		ServiceNodeMaxPauseBlocks:                672,
		ServiceNodesPerSession:                   24,
		FishermanMinimumStake:                    types.BigIntToString(big.NewInt(15000000000)),
		FishermanMaxChains:                       15,
		FishermanUnstakingBlocks:                 2016,
		FishermanMinimumPauseBlocks:              4,
		FishermanMaxPauseBlocks:                  672,
		ValidatorMinimumStake:                    types.BigIntToString(big.NewInt(15000000000)),
		ValidatorUnstakingBlocks:                 2016,
		ValidatorMinimumPauseBlocks:              4,
		ValidatorMaxPauseBlocks:                  672,
		ValidatorMaximumMissedBlocks:             5,
		ValidatorMaxEvidenceAgeInBlocks:          8,
		ProposerPercentageOfFees:                 10,
		MissedBlocksBurnPercentage:               1,
		DoubleSignBurnPercentage:                 5,
		MessageDoubleSignFee:                     types.BigIntToString(big.NewInt(10000)),
		MessageSendFee:                           types.BigIntToString(big.NewInt(10000)),
		MessageStakeFishermanFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeFishermanFee:             types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeFishermanFee:               types.BigIntToString(big.NewInt(10000)),
		MessagePauseFishermanFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseFishermanFee:               types.BigIntToString(big.NewInt(10000)),
		MessageFishermanPauseServiceNodeFee:      types.BigIntToString(big.NewInt(10000)),
		MessageTestScoreFee:                      types.BigIntToString(big.NewInt(10000)),
		MessageProveTestScoreFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageStakeAppFee:                       types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeAppFee:                   types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeAppFee:                     types.BigIntToString(big.NewInt(10000)),
		MessagePauseAppFee:                       types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseAppFee:                     types.BigIntToString(big.NewInt(10000)),
		MessageStakeValidatorFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeValidatorFee:             types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeValidatorFee:               types.BigIntToString(big.NewInt(10000)),
		MessagePauseValidatorFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseValidatorFee:               types.BigIntToString(big.NewInt(10000)),
		MessageStakeServiceNodeFee:               types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeServiceNodeFee:           types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeServiceNodeFee:             types.BigIntToString(big.NewInt(10000)),
		MessagePauseServiceNodeFee:               types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseServiceNodeFee:             types.BigIntToString(big.NewInt(10000)),
		MessageChangeParameterFee:                types.BigIntToString(big.NewInt(10000)),
		AclOwner:                                 DefaultParamsOwner.Address(),
		BlocksPerSessionOwner:                    DefaultParamsOwner.Address(),
		AppMinimumStakeOwner:                     DefaultParamsOwner.Address(),
		AppMaxChainsOwner:                        DefaultParamsOwner.Address(),
		AppBaselineStakeRateOwner:                DefaultParamsOwner.Address(),
		AppStakingAdjustmentOwner:                DefaultParamsOwner.Address(),
		AppUnstakingBlocksOwner:                  DefaultParamsOwner.Address(),
		AppMinimumPauseBlocksOwner:               DefaultParamsOwner.Address(),
		AppMaxPausedBlocksOwner:                  DefaultParamsOwner.Address(),
		ServiceNodeMinimumStakeOwner:             DefaultParamsOwner.Address(),
		ServiceNodeMaxChainsOwner:                DefaultParamsOwner.Address(),
		ServiceNodeUnstakingBlocksOwner:          DefaultParamsOwner.Address(),
		ServiceNodeMinimumPauseBlocksOwner:       DefaultParamsOwner.Address(),
		ServiceNodeMaxPausedBlocksOwner:          DefaultParamsOwner.Address(),
		ServiceNodesPerSessionOwner:              DefaultParamsOwner.Address(),
		FishermanMinimumStakeOwner:               DefaultParamsOwner.Address(),
		FishermanMaxChainsOwner:                  DefaultParamsOwner.Address(),
		FishermanUnstakingBlocksOwner:            DefaultParamsOwner.Address(),
		FishermanMinimumPauseBlocksOwner:         DefaultParamsOwner.Address(),
		FishermanMaxPausedBlocksOwner:            DefaultParamsOwner.Address(),
		ValidatorMinimumStakeOwner:               DefaultParamsOwner.Address(),
		ValidatorUnstakingBlocksOwner:            DefaultParamsOwner.Address(),
		ValidatorMinimumPauseBlocksOwner:         DefaultParamsOwner.Address(),
		ValidatorMaxPausedBlocksOwner:            DefaultParamsOwner.Address(),
		ValidatorMaximumMissedBlocksOwner:        DefaultParamsOwner.Address(),
		ValidatorMaxEvidenceAgeInBlocksOwner:     DefaultParamsOwner.Address(),
		ProposerPercentageOfFeesOwner:            DefaultParamsOwner.Address(),
		MissedBlocksBurnPercentageOwner:          DefaultParamsOwner.Address(),
		DoubleSignBurnPercentageOwner:            DefaultParamsOwner.Address(),
		MessageDoubleSignFeeOwner:                DefaultParamsOwner.Address(),
		MessageSendFeeOwner:                      DefaultParamsOwner.Address(),
		MessageStakeFishermanFeeOwner:            DefaultParamsOwner.Address(),
		MessageEditStakeFishermanFeeOwner:        DefaultParamsOwner.Address(),
		MessageUnstakeFishermanFeeOwner:          DefaultParamsOwner.Address(),
		MessagePauseFishermanFeeOwner:            DefaultParamsOwner.Address(),
		MessageUnpauseFishermanFeeOwner:          DefaultParamsOwner.Address(),
		MessageFishermanPauseServiceNodeFeeOwner: DefaultParamsOwner.Address(),
		MessageTestScoreFeeOwner:                 DefaultParamsOwner.Address(),
		MessageProveTestScoreFeeOwner:            DefaultParamsOwner.Address(),
		MessageStakeAppFeeOwner:                  DefaultParamsOwner.Address(),
		MessageEditStakeAppFeeOwner:              DefaultParamsOwner.Address(),
		MessageUnstakeAppFeeOwner:                DefaultParamsOwner.Address(),
		MessagePauseAppFeeOwner:                  DefaultParamsOwner.Address(),
		MessageUnpauseAppFeeOwner:                DefaultParamsOwner.Address(),
		MessageStakeValidatorFeeOwner:            DefaultParamsOwner.Address(),
		MessageEditStakeValidatorFeeOwner:        DefaultParamsOwner.Address(),
		MessageUnstakeValidatorFeeOwner:          DefaultParamsOwner.Address(),
		MessagePauseValidatorFeeOwner:            DefaultParamsOwner.Address(),
		MessageUnpauseValidatorFeeOwner:          DefaultParamsOwner.Address(),
		MessageStakeServiceNodeFeeOwner:          DefaultParamsOwner.Address(),
		MessageEditStakeServiceNodeFeeOwner:      DefaultParamsOwner.Address(),
		MessageUnstakeServiceNodeFeeOwner:        DefaultParamsOwner.Address(),
		MessagePauseServiceNodeFeeOwner:          DefaultParamsOwner.Address(),
		MessageUnpauseServiceNodeFeeOwner:        DefaultParamsOwner.Address(),
		MessageChangeParameterFeeOwner:           DefaultParamsOwner.Address(),
	}
}
