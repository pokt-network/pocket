package types

var poolFriendlyNames map[PoolNames]string

func init() {
	poolFriendlyNames = map[PoolNames]string{
		PoolNames_POOL_NAMES_UNSPECIFIED:        "Unspecified",
		PoolNames_POOL_NAMES_DAO:                "DAO",
		PoolNames_POOL_NAMES_FEE_COLLECTOR:      "FeeCollector",
		PoolNames_POOL_NAMES_APP_STAKE:          "AppStakePool",
		PoolNames_POOL_NAMES_VALIDATOR_STAKE:    "ValidatorStakePool",
		PoolNames_POOL_NAMES_SERVICE_NODE_STAKE: "ServiceNodeStakePool",
	}
}

func (pn PoolNames) FriendlyName() string {
	return poolFriendlyNames[pn]
}
