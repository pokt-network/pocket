package modules

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/debug"
)

type PersistenceModule interface {
	Module
	NewRWContext(height int64) (PersistenceRWContext, error)
	NewReadContext(height int64) (PersistenceReadContext, error)

	// TODO(drewsky): Make this a context function only and do not expose it at the module level.
	//                The reason `Olshansky` originally made it a module level function is because
	//                the module was responsible for maintaining a single write context and assuring
	//                that a second can't be created (or a previous one is cleaned up) but there is
	//                likely a better and cleaner approach that simplifies the interface.
	ResetContext() error
	GetBlockStore() kvstore.KVStore

	// Debugging / development only
	HandleDebugMessage(*debug.DebugMessage) error
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

// NOTE: There's not really a use case for a write only interface,
// but it abstracts and contrasts nicely against the read only context
// TODO (andrew) convert address and public key to string not bytes #149
type PersistenceWriteContext interface {
	// TODO: Simplify the interface (reference - https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces)
	// - Add general purpose methods such as `ActorOperation(enum_actor_type, ...)` which can be use like so: `Insert(FISHERMAN, ...)`
	// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_ADJUSTMENT, ...)`
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error

	Reset() error
	Commit() error
	Release() error

	AppHash() ([]byte, error)

	// Block Operations

	// Indexer Operations
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
	SetPoolAmount(name string, amount string) error

	InsertPool(name string, address []byte, amount string) error // TODO (Andrew) remove address from pool #149

	// Account Operations
	AddAccountAmount(address []byte, amount string) error
	SubtractAccountAmount(address []byte, amount string) error
	SetAccountAmount(address []byte, amount string) error // NOTE: same as (insert)

	// App Operations
	InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateApp(address []byte, maxRelaysToAdd string, amount string, chainsToUpdate []string) error
	DeleteApp(address []byte) error
	SetAppStakeAmount(address []byte, stakeAmount string) error
	SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetAppPauseHeight(address []byte, height int64) error

	// ServiceNode Operations
	InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateServiceNode(address []byte, serviceURL string, amount string, chains []string) error
	DeleteServiceNode(address []byte) error
	SetServiceNodeStakeAmount(address []byte, stakeAmount string) error
	SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetServiceNodePauseHeight(address []byte, height int64) error

	// Fisherman Operations
	InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateFisherman(address []byte, serviceURL string, amount string, chains []string) error
	DeleteFisherman(address []byte) error
	SetFishermanStakeAmount(address []byte, stakeAmount string) error
	SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetFishermanPauseHeight(address []byte, height int64) error

	// Validator Operations
	InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error
	UpdateValidator(address []byte, serviceURL string, amount string) error
	DeleteValidator(address []byte) error
	SetValidatorStakeAmount(address []byte, stakeAmount string) error
	SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetValidatorPauseHeight(address []byte, height int64) error
	SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error
	SetValidatorMissedBlocks(address []byte, missedBlocks int) error

	/* TODO(olshansky): review/revisit this in more details */

	// Param Operations
	InitParams() error
	SetParam(paramName string, value interface{}) error

	// Flag Operations
	InitFlags() error
	SetFlag(paramName string, value interface{}, enabled bool) error
}

type PersistenceReadContext interface {
	GetHeight() (int64, error)

	// Closes the read context
	Close() error

	// Block Queries
	GetLatestBlockHeight() (uint64, error)
	GetBlockHash(height int64) ([]byte, error)
	GetBlocksPerSession(height int64) (int, error)

	// Indexer Queries
	TransactionExists(transactionHash string) (bool, error)

	// Pool Queries

	// Returns "0" if the account does not exist
	GetPoolAmount(name string, height int64) (amount string, err error)
	GetAllPools(height int64) ([]Account, error)

	// Account Queries

	// Returns "0" if the account does not exist
	GetAccountAmount(address []byte, height int64) (string, error)
	GetAllAccounts(height int64) ([]Account, error)

	// App Queries
	GetAllApps(height int64) ([]Actor, error)
	GetAppExists(address []byte, height int64) (exists bool, err error)
	GetAppStakeAmount(height int64, address []byte) (string, error)
	GetAppsReadyToUnstake(height int64, status int) (apps []UnstakingActorI, err error)
	GetAppStatus(address []byte, height int64) (status int, err error)
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// ServiceNode Queries
	GetAllServiceNodes(height int64) ([]Actor, error)
	GetServiceNodeExists(address []byte, height int64) (exists bool, err error)
	GetServiceNodeStakeAmount(height int64, address []byte) (string, error)
	GetServiceNodesReadyToUnstake(height int64, status int) (serviceNodes []UnstakingActorI, err error)
	GetServiceNodeStatus(address []byte, height int64) (status int, err error)
	GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error)
	GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetServiceNodeCount(chain string, height int64) (int, error)
	GetServiceNodesPerSessionAt(height int64) (int, error)

	// Fisherman Queries
	GetAllFishermen(height int64) ([]Actor, error)
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	GetFishermenReadyToUnstake(height int64, status int) (fishermen []UnstakingActorI, err error)
	GetFishermanStatus(address []byte, height int64) (status int, err error)
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Queries
	GetAllValidators(height int64) ([]Actor, error)
	GetValidatorExists(address []byte, height int64) (exists bool, err error)
	GetValidatorStakeAmount(height int64, address []byte) (string, error)
	GetValidatorsReadyToUnstake(height int64, status int) (validators []UnstakingActorI, err error)
	GetValidatorStatus(address []byte, height int64) (status int, err error)
	GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetValidatorMissedBlocks(address []byte, height int64) (int, error)

	/* TODO(olshansky): review/revisit this in more details */

	// Params
	GetIntParam(paramName string, height int64) (int, error)
	GetStringParam(paramName string, height int64) (string, error)
	GetBytesParam(paramName string, height int64) ([]byte, error)

	// Flags
	GetIntFlag(paramName string, height int64) (int, bool, error)
	GetStringFlag(paramName string, height int64) (string, bool, error)
	GetBytesFlag(paramName string, height int64) ([]byte, bool, error)
}
