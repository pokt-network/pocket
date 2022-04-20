package types

// TODO(team): Consider refactoring PoolNames and statuses to an enum
// with appropriate enum <-> string mappers where appropriate.
// This can make it easier to track all the different states
// available.
const (
	ServiceNodeStakePoolName = "SERVICE_NODE_STAKE_POOL"
	AppStakePoolName         = "APP_STAKE_POOL"
	ValidatorStakePoolName   = "VALIDATOR_STAKE_POOL"
	FishermanStakePoolName   = "FISHERMAN_STAKE_POOL"
	DAOPoolName              = "DAO_POOL"
	FeePoolName              = "FEE_POOL"
	UnstakingStatus          = 1
	StakedStatus             = 2
)