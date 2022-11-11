package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/persistence_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/debug"
)

type PersistenceModule interface {
	Module
	ConfigurableModule
	GenesisDependentModule

	// Context operations
	NewRWContext(height int64) (PersistenceRWContext, error)
	NewReadContext(height int64) (PersistenceReadContext, error)
	ReleaseWriteContext() error // The module can maintain many read contexts, but only one write context can exist at a time

	// BlockStore operations
	GetBlockStore() kvstore.KVStore
	NewWriteContext() PersistenceRWContext

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

// REFACTOR: Simplify the interface
// - Add general purpose methods such as `ActorOperation(enum_actor_type, ...)` which can be use like so: `Insert(FISHERMAN, ...)`
// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_ADJUSTMENT, ...)`
// - Reference: https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces

// TOD (#149): convert address and public key to string from bytes
// NOTE: There's not really a use case for a write only interface, but it abstracts and contrasts nicely against the read only context
type PersistenceWriteContext interface {
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error
	Release() error

	// Commits the current context (height, hash, transactions, etc...) to finality.
	Commit(quorumCert []byte) error

	// Indexer Operations

	// Block Operations
	SetProposalBlock(blockHash string, proposerAddr []byte, quorumCert []byte, transactions [][]byte) error
	GetLatestBlockTxs() [][]byte              // Returns the transactions set by `SetProposalBlock`
	ComputeAppHash() ([]byte, error)          // Update the merkle trees, computes the new state hash, and returns in
	IndexTransaction(txResult TxResult) error // DISCUSS_IN_THIS_COMMIT: How can we remove `TxResult` from the public interface?

	// Pool Operations
	AddPoolAmount(name string, amount string) error
	SubtractPoolAmount(name string, amount string) error
	SetPoolAmount(name string, amount string) error
	InsertPool(name string, address []byte, amount string) error // TODO(#149): remove address from pool

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

	// ServiceNode Operations
	InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int32, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateServiceNode(address []byte, serviceURL string, amount string, chains []string) error
	SetServiceNodeStakeAmount(address []byte, stakeAmount string) error
	SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetServiceNodePauseHeight(address []byte, height int64) error

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
	SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetValidatorPauseHeight(address []byte, height int64) error
	SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error
	SetValidatorMissedBlocks(address []byte, missedBlocks int) error

	// Param Operations
	InitParams() error
	SetParam(paramName string, value interface{}) error

	// Flag Operations
	InitFlags() error
	SetFlag(paramName string, value interface{}, enabled bool) error
}

type PersistenceReadContext interface {
	// Context Operations
	GetHeight() (int64, error) // Returns the height of the context
	Close() error              // Closes the read context

	// Block Queries
	GetLatestBlockHeight() (uint64, error)         // Returns the height of the latest block in the persistence layer
	GetBlockHash(height int64) ([]byte, error)     // Returns the app hash corresponding to the height provides
	GetLatestProposerAddr() []byte                 // Returns the proposer set via `SetProposalBlock`
	GetBlocksPerSession(height int64) (int, error) // TECHDEBT(#286): Deprecate this method

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
	GetAppsReadyToUnstake(height int64, status int32) (apps []IUnstakingActor, err error)
	GetAppStatus(address []byte, height int64) (status int32, err error)
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// ServiceNode Queries
	GetAllServiceNodes(height int64) ([]Actor, error)
	GetServiceNodeExists(address []byte, height int64) (exists bool, err error)
	GetServiceNodeStakeAmount(height int64, address []byte) (string, error)
	GetServiceNodesReadyToUnstake(height int64, status int32) (serviceNodes []IUnstakingActor, err error)
	GetServiceNodeStatus(address []byte, height int64) (status int32, err error)
	GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error)
	GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetServiceNodeCount(chain string, height int64) (int, error)
	GetServiceNodesPerSessionAt(height int64) (int, error) // TECHDEBT(#286): Deprecate this method

	// Fisherman Queries
	GetAllFishermen(height int64) ([]Actor, error)
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	GetFishermenReadyToUnstake(height int64, status int32) (fishermen []IUnstakingActor, err error)
	GetFishermanStatus(address []byte, height int64) (status int32, err error)
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Queries
	GetAllValidators(height int64) ([]Actor, error)
	GetValidatorExists(address []byte, height int64) (exists bool, err error)
	GetValidatorStakeAmount(height int64, address []byte) (string, error)
	GetValidatorsReadyToUnstake(height int64, status int32) (validators []IUnstakingActor, err error)
	GetValidatorStatus(address []byte, height int64) (status int32, err error)
	GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetValidatorMissedBlocks(address []byte, height int64) (int, error)

	// Params
	GetIntParam(paramName string, height int64) (int, error)
	GetStringParam(paramName string, height int64) (string, error)
	GetBytesParam(paramName string, height int64) ([]byte, error)

	// Flags
	GetIntFlag(paramName string, height int64) (int, bool, error)
	GetStringFlag(paramName string, height int64) (string, bool, error)
	GetBytesFlag(paramName string, height int64) ([]byte, bool, error)
}
