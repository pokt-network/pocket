package modules

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

type PersistenceModule interface {
	Module

	NewContext(height int64) (PersistenceContext, error)
	GetCommitDB() *memdb.DB
	GetBlockStore() kvstore.KVStore

	// Debugging / development only
	HandleDebugMessage(*types.DebugMessage) error
}

// Interface defining the context within which the node can operate with the persistence layer.
// Operations in the context of a PersistenceContext are isolated from other operations and
// other persistence contexts until committed, enabling parallelizability along other operations.

// By design, the interface is made very verbose and explicit. This highlights the fact that Pocket
// is an application specific blockchain and improves readability throughout the rest of the codebase
// by limiting the use of abstractions.

// TODO: Simplify the interface (reference - https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces)
// - Add general purpose methods such as `ActorOperation(enum_actor_type, ...)` which can be use like so: `Insert(FISHERMAN, ...)`
// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_ADJUSTMENT, ...)`
type PersistenceContext interface {
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error

	Reset() error
	Commit() error
	Release() // IMPROVE: Return an error?

	AppHash() ([]byte, error)
	GetHeight() (int64, error)

	// Block Operations
	GetLatestBlockHeight() (int64, error)
	GetBlockHash(height int64) ([]byte, error)
	GetBlocksPerSession(height int64) (int, error)

	// Indexer Operations
	TransactionExists(transactionHash string) (bool, error)
	StoreTransaction(transactionProtoBytes []byte) error

	// Block Operations
	// TODO_TEMPORARY: Including two functions for the SQL and KV Store as an interim solution
	//                 until we include the schema as part of the SQL Store because persistence
	//                 currently has no access to the protobuf schema which is the source of truth.
	StoreBlock(blockProtoBytes []byte) error                                              // Store the block in the KV Store
	InsertBlock(height uint64, hash string, proposerAddr []byte, quorumCert []byte) error // Writes the block in the SQL database

	// Pool Operations
	AddPoolAmount(name string, amount string) error
	SubtractPoolAmount(name string, amount string) error
	GetPoolAmount(name string, height int64) (amount string, err error)
	SetPoolAmount(name string, amount string) error

	InsertPool(name string, address []byte, amount string) error

	// Account Operations
	AddAccountAmount(address []byte, amount string) error
	SubtractAccountAmount(address []byte, amount string) error
	GetAccountAmount(address []byte, height int64) (string, error)
	SetAccountAmount(address []byte, amount string) error // TECHDEBT(team): Delete this function

	// App Operations
	GetAppExists(address []byte, height int64) (exists bool, err error)
	InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedAmount string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateApp(address []byte, maxRelays string, stakedAmount string, chainsToUpdate []string) error
	DeleteApp(address []byte) error
	GetAppStakeAmount(height int64, address []byte) (string, error)
	SetAppStakeAmount(address []byte, stakeAmount string) error
	GetAppsReadyToUnstake(height int64, status int) (apps []*types.UnstakingActor, err error)
	GetAppStatus(address []byte, height int64) (status int, err error)
	SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetAppPauseHeight(address []byte, height int64) error
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// ServiceNode Operations
	GetServiceNodeExists(address []byte, height int64) (exists bool, err error)
	InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedAmount string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateServiceNode(address []byte, serviceURL string, stakedAmount string, chains []string) error
	DeleteServiceNode(address []byte) error
	GetServiceNodeStakeAmount(height int64, address []byte) (string, error)
	SetServiceNodeStakeAmount(address []byte, stakeAmount string) error
	GetServiceNodesReadyToUnstake(height int64, status int) (serviceNodes []*types.UnstakingActor, err error)
	GetServiceNodeStatus(address []byte, height int64) (status int, err error)
	SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error)
	SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetServiceNodePauseHeight(address []byte, height int64) error
	GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error)

	GetServiceNodeCount(chain string, height int64) (int, error)
	GetServiceNodesPerSessionAt(height int64) (int, error)

	// Fisherman Operations
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedAmount string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateFisherman(address []byte, serviceURL string, stakedAmount string, chains []string) error
	DeleteFisherman(address []byte) error
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	SetFishermanStakeAmount(address []byte, stakeAmount string) error
	GetFishermenReadyToUnstake(height int64, status int) (fishermen []*types.UnstakingActor, err error)
	GetFishermanStatus(address []byte, height int64) (status int, err error)
	SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetFishermanPauseHeight(address []byte, height int64) error
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Operations
	GetValidatorExists(address []byte, height int64) (exists bool, err error)
	InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedAmount string, pausedHeight int64, unstakingHeight int64) error
	UpdateValidator(address []byte, serviceURL string, stakedAmount string) error
	DeleteValidator(address []byte) error
	GetValidatorStakeAmount(height int64, address []byte) (string, error)
	SetValidatorStakeAmount(address []byte, stakeAmount string) error
	GetValidatorsReadyToUnstake(height int64, status int) (validators []*types.UnstakingActor, err error)
	GetValidatorStatus(address []byte, height int64) (status int, err error)
	SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error)
	SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetValidatorPauseHeight(address []byte, height int64) error
	GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error)

	SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error

	SetValidatorMissedBlocks(address []byte, missedBlocks int) error
	GetValidatorMissedBlocks(address []byte, height int64) (int, error)

	/* TODO(olshansky): review/revisit this in more details */

	// Params
	InitParams() error

	GetIntParam(paramName string, height int64) (int, error)
	GetStringParam(paramName string, height int64) (string, error)
	GetBytesParam(paramName string, height int64) ([]byte, error)
	SetParam(paramName string, value interface{}) error

	// Flags
	InitFlags() error

	GetFlag(flagName string, height int64) (bool, error)
	SetFlag(flagName string, value bool) error
}
