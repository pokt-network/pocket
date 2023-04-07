package types

var poolFriendlyNames map[Pools]string
var poolAddresses map[Pools][]byte

func init() {
	poolFriendlyNames = map[Pools]string{
		Pools_POOLS_UNSPECIFIED:     "",
		Pools_POOLS_DAO:             "DAO",
		Pools_POOLS_FEE_COLLECTOR:   "FeeCollector",
		Pools_POOLS_APP_STAKE:       "AppStakePool",
		Pools_POOLS_VALIDATOR_STAKE: "ValidatorStakePool",
		Pools_POOLS_SERVICER_STAKE:  "ServicerStakePool",
		Pools_POOLS_FISHERMAN_STAKE: "FishermanStakePool",
	}

	// poolAddresses is a map of pools to their addresses. This is to avoid using the hack of using the pool name as the address
	poolAddresses = map[Pools][]byte{
		Pools_POOLS_UNSPECIFIED:     []byte(""),
		Pools_POOLS_DAO:             []byte("44414f0000000000000000000000000000000000"),
		Pools_POOLS_FEE_COLLECTOR:   []byte("466565436f6c6c6563746f720000000000000000"),
		Pools_POOLS_APP_STAKE:       []byte("4170705374616b65506f6f6c0000000000000000"),
		Pools_POOLS_VALIDATOR_STAKE: []byte("56616c696461746f725374616b65506f6f6c0000"),
		Pools_POOLS_SERVICER_STAKE:  []byte("53657276696365725374616b65506f6f6c000000"),
		Pools_POOLS_FISHERMAN_STAKE: []byte("4669736865726d616e5374616b65506f6f6c0000"),
	}
}

func (pn Pools) FriendlyName() string {
	return poolFriendlyNames[pn]
}

func (pn Pools) Address() []byte {
	return poolAddresses[pn]
}
