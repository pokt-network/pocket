package types

// TODO(team): Consider refactoring PoolNames and statuses to an enum
// with appropriate enum <-> string mappers where appropriate.
// This can make it easier to track all the different states
// available.
const (
	UnstakingStatus = 1
	StakedStatus    = 2
)
