package savepoints

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

var _ modules.PersistenceReadContext = &Savepoint{}

type Savepoint struct{}

// Close implements modules.PersistenceReadContext
func (*Savepoint) Close() error {
	panic("unimplemented")
}

// GetAccountAmount implements modules.PersistenceReadContext
func (*Savepoint) GetAccountAmount(address []byte, height int64) (string, error) {
	panic("unimplemented")
}

// GetAllAccounts implements modules.PersistenceReadContext
func (*Savepoint) GetAllAccounts(height int64) ([]*coreTypes.Account, error) {
	panic("unimplemented")
}

// GetAllApps implements modules.PersistenceReadContext
func (*Savepoint) GetAllApps(height int64) ([]*coreTypes.Actor, error) {
	panic("unimplemented")
}

// GetAllFishermen implements modules.PersistenceReadContext
func (*Savepoint) GetAllFishermen(height int64) ([]*coreTypes.Actor, error) {
	panic("unimplemented")
}

// GetAllPools implements modules.PersistenceReadContext
func (*Savepoint) GetAllPools(height int64) ([]*coreTypes.Account, error) {
	panic("unimplemented")
}

// GetAllServicers implements modules.PersistenceReadContext
func (*Savepoint) GetAllServicers(height int64) ([]*coreTypes.Actor, error) {
	panic("unimplemented")
}

// GetAllStakedActors implements modules.PersistenceReadContext
func (*Savepoint) GetAllStakedActors(height int64) ([]*coreTypes.Actor, error) {
	panic("unimplemented")
}

// GetAllValidators implements modules.PersistenceReadContext
func (*Savepoint) GetAllValidators(height int64) ([]*coreTypes.Actor, error) {
	panic("unimplemented")
}

// GetAppExists implements modules.PersistenceReadContext
func (*Savepoint) GetAppExists(address []byte, height int64) (exists bool, err error) {
	panic("unimplemented")
}

// GetAppOutputAddress implements modules.PersistenceReadContext
func (*Savepoint) GetAppOutputAddress(operator []byte, height int64) (output []byte, err error) {
	panic("unimplemented")
}

// GetAppPauseHeightIfExists implements modules.PersistenceReadContext
func (*Savepoint) GetAppPauseHeightIfExists(address []byte, height int64) (int64, error) {
	panic("unimplemented")
}

// GetAppStakeAmount implements modules.PersistenceReadContext
func (*Savepoint) GetAppStakeAmount(height int64, address []byte) (string, error) {
	panic("unimplemented")
}

// GetAppStatus implements modules.PersistenceReadContext
func (*Savepoint) GetAppStatus(address []byte, height int64) (status int32, err error) {
	panic("unimplemented")
}

// GetAppsReadyToUnstake implements modules.PersistenceReadContext
func (*Savepoint) GetAppsReadyToUnstake(height int64, status int32) (apps []*moduleTypes.UnstakingActor, err error) {
	panic("unimplemented")
}

// GetBlockHash implements modules.PersistenceReadContext
func (*Savepoint) GetBlockHash(height int64) (string, error) {
	panic("unimplemented")
}

// GetBytesFlag implements modules.PersistenceReadContext
func (*Savepoint) GetBytesFlag(paramName string, height int64) ([]byte, bool, error) {
	panic("unimplemented")
}

// GetBytesParam implements modules.PersistenceReadContext
func (*Savepoint) GetBytesParam(paramName string, height int64) ([]byte, error) {
	panic("unimplemented")
}

// GetFishermanExists implements modules.PersistenceReadContext
func (*Savepoint) GetFishermanExists(address []byte, height int64) (exists bool, err error) {
	panic("unimplemented")
}

// GetFishermanOutputAddress implements modules.PersistenceReadContext
func (*Savepoint) GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error) {
	panic("unimplemented")
}

// GetFishermanPauseHeightIfExists implements modules.PersistenceReadContext
func (*Savepoint) GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error) {
	panic("unimplemented")
}

// GetFishermanStakeAmount implements modules.PersistenceReadContext
func (*Savepoint) GetFishermanStakeAmount(height int64, address []byte) (string, error) {
	panic("unimplemented")
}

// GetFishermanStatus implements modules.PersistenceReadContext
func (*Savepoint) GetFishermanStatus(address []byte, height int64) (status int32, err error) {
	panic("unimplemented")
}

// GetFishermenReadyToUnstake implements modules.PersistenceReadContext
func (*Savepoint) GetFishermenReadyToUnstake(height int64, status int32) (fishermen []*moduleTypes.UnstakingActor, err error) {
	panic("unimplemented")
}

// GetHeight implements modules.PersistenceReadContext
func (*Savepoint) GetHeight() (int64, error) {
	panic("unimplemented")
}

// GetIntFlag implements modules.PersistenceReadContext
func (*Savepoint) GetIntFlag(paramName string, height int64) (int, bool, error) {
	panic("unimplemented")
}

// GetIntParam implements modules.PersistenceReadContext
func (*Savepoint) GetIntParam(paramName string, height int64) (int, error) {
	panic("unimplemented")
}

// GetMaximumBlockHeight implements modules.PersistenceReadContext
func (*Savepoint) GetMaximumBlockHeight() (uint64, error) {
	panic("unimplemented")
}

// GetMinimumBlockHeight implements modules.PersistenceReadContext
func (*Savepoint) GetMinimumBlockHeight() (uint64, error) {
	panic("unimplemented")
}

// GetParameter implements modules.PersistenceReadContext
func (*Savepoint) GetParameter(paramName string, height int64) (any, error) {
	panic("unimplemented")
}

// GetPoolAmount implements modules.PersistenceReadContext
func (*Savepoint) GetPoolAmount(name string, height int64) (amount string, err error) {
	panic("unimplemented")
}

// GetServicerCount implements modules.PersistenceReadContext
func (*Savepoint) GetServicerCount(chain string, height int64) (int, error) {
	panic("unimplemented")
}

// GetServicerExists implements modules.PersistenceReadContext
func (*Savepoint) GetServicerExists(address []byte, height int64) (exists bool, err error) {
	panic("unimplemented")
}

// GetServicerOutputAddress implements modules.PersistenceReadContext
func (*Savepoint) GetServicerOutputAddress(operator []byte, height int64) (output []byte, err error) {
	panic("unimplemented")
}

// GetServicerPauseHeightIfExists implements modules.PersistenceReadContext
func (*Savepoint) GetServicerPauseHeightIfExists(address []byte, height int64) (int64, error) {
	panic("unimplemented")
}

// GetServicerStakeAmount implements modules.PersistenceReadContext
func (*Savepoint) GetServicerStakeAmount(height int64, address []byte) (string, error) {
	panic("unimplemented")
}

// GetServicerStatus implements modules.PersistenceReadContext
func (*Savepoint) GetServicerStatus(address []byte, height int64) (status int32, err error) {
	panic("unimplemented")
}

// GetServicersReadyToUnstake implements modules.PersistenceReadContext
func (*Savepoint) GetServicersReadyToUnstake(height int64, status int32) (servicers []*moduleTypes.UnstakingActor, err error) {
	panic("unimplemented")
}

// GetStringFlag implements modules.PersistenceReadContext
func (*Savepoint) GetStringFlag(paramName string, height int64) (string, bool, error) {
	panic("unimplemented")
}

// GetStringParam implements modules.PersistenceReadContext
func (*Savepoint) GetStringParam(paramName string, height int64) (string, error) {
	panic("unimplemented")
}

// GetValidatorExists implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorExists(address []byte, height int64) (exists bool, err error) {
	panic("unimplemented")
}

// GetValidatorMissedBlocks implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorMissedBlocks(address []byte, height int64) (int, error) {
	panic("unimplemented")
}

// GetValidatorOutputAddress implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error) {
	panic("unimplemented")
}

// GetValidatorPauseHeightIfExists implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error) {
	panic("unimplemented")
}

// GetValidatorStakeAmount implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorStakeAmount(height int64, address []byte) (string, error) {
	panic("unimplemented")
}

// GetValidatorStatus implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorStatus(address []byte, height int64) (status int32, err error) {
	panic("unimplemented")
}

// GetValidatorsReadyToUnstake implements modules.PersistenceReadContext
func (*Savepoint) GetValidatorsReadyToUnstake(height int64, status int32) (validators []*moduleTypes.UnstakingActor, err error) {
	panic("unimplemented")
}
