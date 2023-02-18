package types

var poolFriendlyNames map[Pools]string

func init() {
	poolFriendlyNames = map[Pools]string{
		Pools_POOLS_UNSPECIFIED:        "Unspecified",
		Pools_POOLS_DAO:                "DAO",
		Pools_POOLS_FEE_COLLECTOR:      "FeeCollector",
		Pools_POOLS_APP_STAKE:          "AppStakePool",
		Pools_POOLS_VALIDATOR_STAKE:    "ValidatorStakePool",
		Pools_POOLS_SERVICE_NODE_STAKE: "ServicerStakePool",
	}
}

func (pn Pools) FriendlyName() string {
	return poolFriendlyNames[pn]
}
