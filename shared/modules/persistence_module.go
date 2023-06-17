package modules

//go:generate mockgen -destination=./mocks/persistence_module_mock.go github.com/pokt-network/pocket/shared/modules PersistenceModule,PersistenceRWContext,PersistenceReadContext,PersistenceWriteContext,PersistenceLocalContext

import (
	"math/big"

	"github.com/pokt-network/pocket/persistence/blockstore"
	"github.com/pokt-network/pocket/persistence/indexer"
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
	// TODO(#406): removing height from "NewReadContext" input and passing it to specific methods seems a better choice.
	//	This could prevent confusion when retrieving the value of a parameter for a height less than the current height,
	//		e.g. when getting the App Token sessions multiplier for the starting height of a session.
	NewReadContext(height int64) (PersistenceReadContext, error)
	ReleaseWriteContext() error // The module can maintain many read contexts, but only one write context can exist at a time

	// BlockStore operations
	GetBlockStore() blockstore.BlockStore
	NewWriteContext() PersistenceRWContext

	// Indexer operations
	GetTxIndexer() indexer.TxIndexer
	TransactionExists(transactionHash string) (bool, error)

	// Debugging / development only
	HandleDebugMessage(*messaging.DebugMessage) error

	// GetLocalContext returns a local persistence context that can be used to store/retrieve node-specific, i.e. off-chain, data
	// The module can maintain a single local context for both read and write operations: subsequent calls to GetLocalContext return
	// the same local context.
	GetLocalContext() (PersistenceLocalContext, error)
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
// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_, ...)`
// - Reference: https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces

// TECHDEBT:
//   - Decouple functions that can be split into two or more independent behaviours (e.g. `SetAppStatusAndUnstakingHeightIfPausedBefore`)
//   - Rename `Unstaking` to `Unbonding` where appropriate
//   - convert address and public key to string from bytes
//   - Remove `height` from all write context functions because it should only write at the height it was initiated in

// PersistenceWriteContext has no use-case independent of `PersistenceRWContext`, but is a useful abstraction
type PersistenceWriteContext interface {
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error
	Release()

	// Commits (and releases) the current context to disk (i.e. finality).
	Commit(proposerAddr, quorumCert []byte) error

	// Indexer Operations

	// ComputeStateHash updates the merkle trees, computes the new state hash (i.e. state commitment)
	// if the context is committed.
	ComputeStateHash() (string, error)

	// Indexes the transaction using several different keys (for lookup purposes) in the key-value store
	// that backs the transaction merkle tree.
	IndexTransaction(idxTx *coreTypes.IndexedTransaction) error

	// Pool Operations
	AddPoolAmount(address []byte, amount string) error
	SubtractPoolAmount(address []byte, amount string) error
	SetPoolAmount(address []byte, amount string) error
	InsertPool(address []byte, amount string) error

	// Account Operations
	AddAccountAmount(address []byte, amount string) error
	SubtractAccountAmount(address []byte, amount string) error
	SetAccountAmount(address []byte, amount string) error // NOTE: same as (insert)

	// App Operations
	InsertApp(address []byte, publicKey []byte, output []byte, paused bool, status int32, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateApp(address []byte, amount string, chainsToUpdate []string) error
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
	SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int32) error
	SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int32) error
	SetValidatorPauseHeight(address []byte, height int64) error
	SetValidatorMissedBlocks(address []byte, missedBlocks int) error

	// Param Operations
	InitGenesisParams(params *genesis.Params) error
	SetParam(paramName string, value any) error

	// Flag Operations
	InitFlags() error
	SetFlag(paramName string, value any, enabled bool) error

	RecordRelayService(applicationAddress string, key []byte, relay *coreTypes.Relay, response *coreTypes.RelayResponse) error
}

type PersistenceReadContext interface {
	// Context Operations
	// TECHDEBT: Remove this function since read contexts are height agnostic - it's an accessor to the state of the blockchain at any height.
	GetHeight() (int64, error) // Returns the height of the context
	Release()                  // Releases the read context

	// Version queries
	GetVersionAtHeight(height int64) (string, error) // TODO: Implement this

	// Supported Chains Queries
	GetSupportedChains(height int64) ([]string, error) // TODO: Implement this

	// CONSOLIDATE: BlockHash / AppHash / StateHash
	// Block Queries
	GetMaximumBlockHeight() (uint64, error)    // Returns the height of the latest block in the persistence layer
	GetMinimumBlockHeight() (uint64, error)    // Returns the min block height in the persistence layer
	GetBlockHash(height int64) (string, error) // Returns the app hash corresponding to the height provided

	// Pool Queries

	// Returns "0" if the account does not exist
	GetPoolAmount(address []byte, height int64) (amount string, err error)
	GetAllPools(height int64) ([]*coreTypes.Account, error)

	// Account Queries

	// Returns "0" if the account does not exist
	GetAccountAmount(address []byte, height int64) (string, error)
	GetAllAccounts(height int64) ([]*coreTypes.Account, error)

	// Actor Queries
	GetActor(actorType coreTypes.ActorType, address []byte, height int64) (*coreTypes.Actor, error)

	// App Queries
	GetApp(address []byte, height int64) (*coreTypes.Actor, error)
	GetAllApps(height int64) ([]*coreTypes.Actor, error)
	GetAppExists(address []byte, height int64) (exists bool, err error)
	GetAppStakeAmount(height int64, address []byte) (string, error)
	GetAppsReadyToUnstake(height int64, status int32) (apps []*moduleTypes.UnstakingActor, err error)
	GetAppStatus(address []byte, height int64) (status int32, err error)
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Servicer Queries
	GetServicer(address []byte, height int64) (*coreTypes.Actor, error)
	GetAllServicers(height int64) ([]*coreTypes.Actor, error)
	GetServicerExists(address []byte, height int64) (exists bool, err error)
	GetServicerStakeAmount(height int64, address []byte) (string, error)
	GetServicersReadyToUnstake(height int64, status int32) (servicers []*moduleTypes.UnstakingActor, err error)
	GetServicerStatus(address []byte, height int64) (status int32, err error)
	GetServicerPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetServicerOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetServicerCount(chain string, height int64) (int, error)

	// Fisherman Queries
	GetFisherman(address []byte, height int64) (*coreTypes.Actor, error)
	GetAllFishermen(height int64) ([]*coreTypes.Actor, error)
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	GetFishermenReadyToUnstake(height int64, status int32) (fishermen []*moduleTypes.UnstakingActor, err error)
	GetFishermanStatus(address []byte, height int64) (status int32, err error)
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Queries
	GetValidator(address []byte, height int64) (*coreTypes.Actor, error)
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
	GetAllParams() ([][]string, error)

	// Flags
	GetIntFlag(paramName string, height int64) (int, bool, error)
	GetStringFlag(paramName string, height int64) (string, bool, error)
	GetBytesFlag(paramName string, height int64) ([]byte, bool, error)
}

// PersistenceLocalContext defines the set of operations specific to local persistence.
//
//	This context should be used for node-specific data, e.g. records of served relays.
//	This is in contrast to PersistenceRWContext which should be used to store on-chain data.
type PersistenceLocalContext interface {
	// StoreServicedRelay stores record of a serviced relay and its response in the local context.
	// The stored service relays will be used to:
	//	a) check the number of tokens used per session, and
	//	b) prepare claim/proof messages once the session is over
	// The "relayDigest" and "relayReqResBytes" parameters will be used as key and leaf contents in the constructed SMT, respectively.
	StoreServicedRelay(session *coreTypes.Session, relayDigest, relayReqResBytes []byte) error
	// GetSessionTokensUsed returns the number of tokens that have been used for the provided session.
	//    It returns the count of tokens used by the servicer instance
	//	for the application associated with the session
	GetSessionTokensUsed(*coreTypes.Session) (*big.Int, error)
}
