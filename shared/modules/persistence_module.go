package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/persistence_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime/genesis"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
)

const PersistenceModuleName = "persistence"

type PersistenceModule interface {
	Module

	// Context operations
	NewRWContext(height int64) (PersistenceRWContext, error)
	NewReadContext(height int64) (PersistenceReadContext, error)
	ReleaseWriteContext() error // The module can maintain many read contexts, but only one write context can exist at a time

	// BlockStore operations
	GetBlockStore() kvstore.KVStore
	NewWriteContext() PersistenceRWContext

	// Indexer Queries
	TransactionExists(transactionHash string) (bool, error)

	// Debugging / development only
	HandleDebugMessage(*messaging.DebugMessage) error
}

// Interface defining the context within which the node can operate with the persistence layer.
// Operations in the context of a PersistenceContext are isolated from other operations and
// other persistence contexts until committed, enabling parallelizability along other operations.

// By design, the interface is made very verbose and explicit. This highlights the fact that Pocket
// is an application specific blockchain and improves readability throughout the rest of the codebase
// by limiting the use of abstractions.

// NOTE: Only the Utility Module should use the RW context
type PersistenceRWContext interface {
	PersistenceReadContext
	PersistenceWriteContext
}

// REFACTOR: Simplify the interface
// - Add general purpose methods such as `ActorOperation(enum_actor_type, ...)` which can be use like so: `Insert(FISHERMAN, ...)`
// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_ADJUSTMENT, ...)`
// - Reference: https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces

// TECHDEBT: convert address and public key to string from bytes
// NOTE: There's not really a use case for a write only interface, but it abstracts and contrasts nicely against the read only context
type PersistenceWriteContext interface {
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error
	Release() error

	// Commits the current context (height, hash, transactions, etc...) to disk (i.e. finality).
	Commit(proposerAddr, quorumCert []byte) error

	// Indexer Operations

	// Block Operations
	ComputeStateHash() (string, error)        // Update the merkle trees, computes the new state hash, and returns it
	IndexTransaction(txResult TxResult) error // TODO(#361): Look into an approach to remove `TxResult` from shared interfaces

	// Pool Operations
	AddPoolAmount(name string, amount string) error
	SubtractPoolAmount(name string, amount string) error
	SetPoolAmount(name string, amount string) error
	InsertPool(name string, amount string) error

	// Account Operations
	AddAccountAmount(address []byte, amount string) error
	SubtractAccountAmount(address []byte, amount string) error
	SetAccountAmount(address []byte, amount string) error // NOTE: same as (insert)

	// App Operations
	InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int32, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateApp(address []byte, maxRelaysToAdd string, amount string, chainsToUpdate []string) error
	SetAppStakeAmount(address []byte, stakeAmount string) error
	SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetAppPauseHeight(address []byte, height int64) error

	// Servicer Operations
	InsertServicer(address []byte, publicKey []byte, output []byte, paused bool, status int32, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateServicer(address []byte, serviceURL string, amount string, chains []string) error
	SetServicerStakeAmount(address []byte, stakeAmount string) error
	SetServicerUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetServicerStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetServicerPauseHeight(address []byte, height int64) error

	// Fisherman Operations
	InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int32, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateFisherman(address []byte, serviceURL string, amount string, chains []string) error
	SetFishermanStakeAmount(address []byte, stakeAmount string) error
	SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetFishermanPauseHeight(address []byte, height int64) error

	// Validator Operations
	InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int32, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error
	UpdateValidator(address []byte, serviceURL string, amount string) error
	SetValidatorStakeAmount(address []byte, stakeAmount string) error
	// IMPROVE: Decouple and/or rename these functions
	SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetValidatorPauseHeight(address []byte, height int64) error
	SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error
	SetValidatorMissedBlocks(address []byte, missedBlocks int) error

	// Param Operations
	InitGenesisParams(params *genesis.Params) error
	SetParam(paramName string, value any) error

	// Flag Operations
	InitFlags() error
	SetFlag(paramName string, value any, enabled bool) error
}

type PersistenceReadContext interface {
	// Context Operations
	GetHeight() (int64, error) // Returns the height of the context
	Close() error              // Closes the read context

	// CONSOLIDATE: BlockHash / AppHash / StateHash
	// Block Queries
	GetMaximumBlockHeight() (uint64, error)    // Returns the height of the latest block in the persistence layer
	GetMinimumBlockHeight() (uint64, error)    // Returns the min block height in the persistence layer
	GetBlockHash(height int64) (string, error) // Returns the app hash corresponding to the height provided

	// Pool Queries

	// Returns "0" if the account does not exist
	GetPoolAmount(name string, height int64) (amount string, err error)
	GetAllPools(height int64) ([]*coreTypes.Account, error)

	// Account Queries

	// Returns "0" if the account does not exist
	GetAccountAmount(address []byte, height int64) (string, error)
	GetAllAccounts(height int64) ([]*coreTypes.Account, error)

	// App Queries
	GetAllApps(height int64) ([]*coreTypes.Actor, error)
	GetAppExists(address []byte, height int64) (exists bool, err error)
	GetAppStakeAmount(height int64, address []byte) (string, error)
	GetAppsReadyToUnstake(height int64, status int32) (apps []*moduleTypes.UnstakingActor, err error)
	GetAppStatus(address []byte, height int64) (status int32, err error)
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Servicer Queries
	GetAllServicers(height int64) ([]*coreTypes.Actor, error)
	GetServicerExists(address []byte, height int64) (exists bool, err error)
	GetServicerStakeAmount(height int64, address []byte) (string, error)
	GetServicersReadyToUnstake(height int64, status int32) (servicers []*moduleTypes.UnstakingActor, err error)
	GetServicerStatus(address []byte, height int64) (status int32, err error)
	GetServicerPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetServicerOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetServicerCount(chain string, height int64) (int, error)

	// Fisherman Queries
	GetAllFishermen(height int64) ([]*coreTypes.Actor, error)
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	GetFishermenReadyToUnstake(height int64, status int32) (fishermen []*moduleTypes.UnstakingActor, err error)
	GetFishermanStatus(address []byte, height int64) (status int32, err error)
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Queries
	GetAllValidators(height int64) ([]*coreTypes.Actor, error)
	GetValidatorExists(address []byte, height int64) (exists bool, err error)
	GetValidatorStakeAmount(height int64, address []byte) (string, error)
	GetValidatorsReadyToUnstake(height int64, status int32) (validators []*moduleTypes.UnstakingActor, err error)
	GetValidatorStatus(address []byte, height int64) (status int32, err error)
	GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetValidatorMissedBlocks(address []byte, height int64) (int, error)

	// Actors Queries
	GetAllStakedActors(height int64) ([]*coreTypes.Actor, error)

	// Params
	GetIntParam(paramName string, height int64) (int, error)
	GetStringParam(paramName string, height int64) (string, error)
	GetBytesParam(paramName string, height int64) ([]byte, error)
	GetParameter(paramName string, height int64) (any, error)

	// Flags
	GetIntFlag(paramName string, height int64) (int, bool, error)
	GetStringFlag(paramName string, height int64) (string, bool, error)
	GetBytesFlag(paramName string, height int64) ([]byte, bool, error)
}
