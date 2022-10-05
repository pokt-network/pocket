package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/persistence_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/persistence/kvstore" // Should be moved to shared
	"github.com/pokt-network/pocket/shared/debug"
)

type PersistenceModule interface {
	Module

	// Persistence Context Factory Methods
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

// TODO: Simplify the interface (reference - https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces)
// - Add general purpose methods such as `ActorOperation(enum_actor_type, ...)` which can be use like so: `Insert(FISHERMAN, ...)`
// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_ADJUSTMENT, ...)`

// NOTE: There's not really a use case for a write only interface,
// but it abstracts and contrasts nicely against the read only context
// TODO (andrew) convert address and public key to string not bytes #149
type PersistenceWriteContext interface {
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error

	// DISCUSS: Can we consolidate `Reset` and `Release`
	Reset() error
	Release() error

	// Block / indexer operations
	UpdateAppHash() ([]byte, error)
	// Commits the current context (height, hash, transactions, etc...) to finality.
	Commit(proposerAddr []byte, quorumCert []byte) error
	// Indexes the transaction
	StoreTransaction(transactionProtoBytes []byte) error // Stores a transaction

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
	SetAppStakeAmount(address []byte, stakeAmount string) error
	SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetAppPauseHeight(address []byte, height int64) error

	// ServiceNode Operations
	InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateServiceNode(address []byte, serviceURL string, amount string, chains []string) error
	SetServiceNodeStakeAmount(address []byte, stakeAmount string) error
	SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetServiceNodePauseHeight(address []byte, height int64) error

	// Fisherman Operations
	InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateFisherman(address []byte, serviceURL string, amount string, chains []string) error
	SetFishermanStakeAmount(address []byte, stakeAmount string) error
	SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetFishermanPauseHeight(address []byte, height int64) error

	// Validator Operations
	InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error
	UpdateValidator(address []byte, serviceURL string, amount string) error
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

	// Tree Operations

	// # Option 1:

	UpdateApplicationsTree([]Actor) error
	// UpdateValidatorsTree([]Actor) error
	// UpdateServiceNodesTree([]Actor) error
	// UpdateFishermanTree([]Actor) error
	// Update<FutureActors>Tree([]Actor) error
	// Update<Other>Tree([]Other) error

	// # Option 2:
	// UpdateActorTree(types.ProtocolActorSchema, []Actor) error
	// Update<Other>Tree([]Other) error

	// # Option 3:
	// UpdateApplicationsTree([]Application) error
	// UpdateValidatorsTree([]Validator) error
	// UpdateServiceNodesTree([]ServiceNode) error
	// UpdateFishermanTree([]Fisherman) error
	// Update<FutureActors>Tree([]FutureActor) error
	// Update<Other>Tree([]Other) error
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
	GetAppsReadyToUnstake(height int64, status int) (apps []IUnstakingActor, err error)
	GetAppStatus(address []byte, height int64) (status int, err error)
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// ServiceNode Queries
	GetAllServiceNodes(height int64) ([]Actor, error)
	GetServiceNodeExists(address []byte, height int64) (exists bool, err error)
	GetServiceNodeStakeAmount(height int64, address []byte) (string, error)
	GetServiceNodesReadyToUnstake(height int64, status int) (serviceNodes []IUnstakingActor, err error)
	GetServiceNodeStatus(address []byte, height int64) (status int, err error)
	GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error)
	GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetServiceNodeCount(chain string, height int64) (int, error)
	GetServiceNodesPerSessionAt(height int64) (int, error)

	// Fisherman Queries
	GetAllFishermen(height int64) ([]Actor, error)
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	GetFishermenReadyToUnstake(height int64, status int) (fishermen []IUnstakingActor, err error)
	GetFishermanStatus(address []byte, height int64) (status int, err error)
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Queries
	GetAllValidators(height int64) ([]Actor, error)
	GetValidatorExists(address []byte, height int64) (exists bool, err error)
	GetValidatorStakeAmount(height int64, address []byte) (string, error)
	GetValidatorsReadyToUnstake(height int64, status int) (validators []IUnstakingActor, err error)
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

	// Tree Operations

	// # Option 1:

	// GetApplicationsUpdatedAtHeight(height int64) ([]Actor, error)
	// GetValidatorsUpdatedAtHeight(height int64) ([]Actor, error)
	// GetServiceNodesUpdatedAtHeight(height int64) ([]Actor, error)
	// GetFishermanUpdatedAtHeight(height int64) ([]Actor, error)
	// Get<FutureActor>UpdatedAtHeight(height int64) ([]Actor, error)
	// Get<Other>UpdatedAtHeight(height int64) ([]Actor, error)
	// Update<Other>Tree(height int64) ([]Actor, error)

	// # Option 2:
	// Get<FutureActor>UpdatedAtHeight(types.ProtocolActorSchema, height int64) ([]Actor, error)
	// Get<Other>UpdatedAtHeight(height int64) ([]Other, error)

	// # Option 3:
	// GetApplicationsUpdatedAtHeight(height int64) ([]Application, error)
	// GetValidatorsUpdatedAtHeight(height int64) ([]Validator, error)
	// GetServiceNodesUpdatedAtHeight(height int64) ([]ServiceNode, error)
	// GetFishermanUpdatedAtHeight(height int64) ([]Fisherman, error)
	// Get<FutureActor>UpdatedAtHeight(height int64) ([]FutureActor, error)
	// Get<Other>UpdatedAtHeight(height int64) ([]Other, error)
}
