package modules

import (
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/types"
)

type PersistenceModule interface {
	Module
	NewRWContext(height int64) (PersistenceRWContext, error)
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

// NOTE: Only the Utility Module should use the RW context
type PersistenceRWContext interface {
	PersistenceReadContext
	PersistenceWriteContext
}

// NOTE: There's not really a use case for a write only interface,
// but it abstracts and contrasts nicely against the read only context
type PersistenceWriteContext interface {
	// TODO: Simplify the interface (reference - https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_prefer_single_method_interfaces)
	// - Add general purpose methods such as `ActorOperation(enum_actor_type, ...)` which can be use like so: `Insert(FISHERMAN, ...)`
	// - Use general purpose parameter methods such as `Set(enum_gov_type, ...)` such as `Set(STAKING_ADJUSTMENT, ...)`
	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error

	Reset() error
	Commit() error
	Release() // IMPROVE: Return an error?

	AppHash() ([]byte, error)

	// Block Operations

	// Indexer Operations
	GetBlockHash(height int64) ([]byte, error)
	GetBlocksPerSession() (int, error)

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

	InsertPool(name string, address []byte, amount string) error

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

	SetBlocksPerSession(int) error

	SetParamAppMinimumStake(string) error
	SetMaxAppChains(int) error
	SetBaselineAppStakeRate(int) error
	SetStakingAdjustment(int) error
	SetAppUnstakingBlocks(int) error
	SetAppMinimumPauseBlocks(int) error
	SetAppMaxPausedBlocks(int) error

	SetParamServiceNodeMinimumStake(string) error
	SetServiceNodeMaxChains(int) error
	SetServiceNodeUnstakingBlocks(int) error
	SetServiceNodeMinimumPauseBlocks(int) error
	SetServiceNodeMaxPausedBlocks(int) error
	SetServiceNodesPerSession(int) error

	SetParamFishermanMinimumStake(string) error
	SetFishermanMaxChains(int) error
	SetFishermanUnstakingBlocks(int) error
	SetFishermanMinimumPauseBlocks(int) error
	SetFishermanMaxPausedBlocks(int) error

	SetParamValidatorMinimumStake(string) error
	SetValidatorUnstakingBlocks(int) error
	SetValidatorMinimumPauseBlocks(int) error
	SetValidatorMaxPausedBlocks(int) error
	SetValidatorMaximumMissedBlocks(int) error
	SetProposerPercentageOfFees(int) error
	SetMaxEvidenceAgeInBlocks(int) error
	SetMissedBlocksBurnPercentage(int) error
	SetDoubleSignBurnPercentage(int) error

	SetMessageDoubleSignFee(string) error
	SetMessageSendFee(string) error
	SetMessageStakeFishermanFee(string) error
	SetMessageEditStakeFishermanFee(string) error
	SetMessageUnstakeFishermanFee(string) error
	SetMessagePauseFishermanFee(string) error
	SetMessageUnpauseFishermanFee(string) error
	SetMessageFishermanPauseServiceNodeFee(string) error
	SetMessageTestScoreFee(string) error
	SetMessageProveTestScoreFee(string) error
	SetMessageStakeAppFee(string) error
	SetMessageEditStakeAppFee(string) error
	SetMessageUnstakeAppFee(string) error
	SetMessagePauseAppFee(string) error
	SetMessageUnpauseAppFee(string) error
	SetMessageStakeValidatorFee(string) error
	SetMessageEditStakeValidatorFee(string) error
	SetMessageUnstakeValidatorFee(string) error
	SetMessagePauseValidatorFee(string) error
	SetMessageUnpauseValidatorFee(string) error
	SetMessageStakeServiceNodeFee(string) error
	SetMessageEditStakeServiceNodeFee(string) error
	SetMessageUnstakeServiceNodeFee(string) error
	SetMessagePauseServiceNodeFee(string) error
	SetMessageUnpauseServiceNodeFee(string) error
	SetMessageChangeParameterFee(string) error

	SetMessageDoubleSignFeeOwner([]byte) error
	SetMessageSendFeeOwner([]byte) error
	SetMessageStakeFishermanFeeOwner([]byte) error
	SetMessageEditStakeFishermanFeeOwner([]byte) error
	SetMessageUnstakeFishermanFeeOwner([]byte) error
	SetMessagePauseFishermanFeeOwner([]byte) error
	SetMessageUnpauseFishermanFeeOwner([]byte) error
	SetMessageFishermanPauseServiceNodeFeeOwner([]byte) error
	SetMessageTestScoreFeeOwner([]byte) error
	SetMessageProveTestScoreFeeOwner([]byte) error
	SetMessageStakeAppFeeOwner([]byte) error
	SetMessageEditStakeAppFeeOwner([]byte) error
	SetMessageUnstakeAppFeeOwner([]byte) error
	SetMessagePauseAppFeeOwner([]byte) error
	SetMessageUnpauseAppFeeOwner([]byte) error
	SetMessageStakeValidatorFeeOwner([]byte) error
	SetMessageEditStakeValidatorFeeOwner([]byte) error
	SetMessageUnstakeValidatorFeeOwner([]byte) error
	SetMessagePauseValidatorFeeOwner([]byte) error
	SetMessageUnpauseValidatorFeeOwner([]byte) error
	SetMessageStakeServiceNodeFeeOwner([]byte) error
	SetMessageEditStakeServiceNodeFeeOwner([]byte) error
	SetMessageUnstakeServiceNodeFeeOwner([]byte) error
	SetMessagePauseServiceNodeFeeOwner([]byte) error
	SetMessageUnpauseServiceNodeFeeOwner([]byte) error
	SetMessageChangeParameterFeeOwner([]byte) error

	// ACL Operations
	SetAclOwner(owner []byte) error
	SetBlocksPerSessionOwner(owner []byte) error
	SetMaxAppChainsOwner(owner []byte) error
	SetAppMinimumStakeOwner(owner []byte) error
	SetBaselineAppOwner(owner []byte) error
	SetStakingAdjustmentOwner(owner []byte) error
	SetAppUnstakingBlocksOwner(owner []byte) error
	SetAppMinimumPauseBlocksOwner(owner []byte) error
	SetAppMaxPausedBlocksOwner(owner []byte) error
	SetServiceNodeMinimumStakeOwner(owner []byte) error
	SetMaxServiceNodeChainsOwner(owner []byte) error
	SetServiceNodeUnstakingBlocksOwner(owner []byte) error
	SetServiceNodeMinimumPauseBlocksOwner(owner []byte) error
	SetServiceNodeMaxPausedBlocksOwner(owner []byte) error
	SetFishermanMinimumStakeOwner(owner []byte) error
	SetMaxFishermanChainsOwner(owner []byte) error
	SetFishermanUnstakingBlocksOwner(owner []byte) error
	SetFishermanMinimumPauseBlocksOwner(owner []byte) error
	SetFishermanMaxPausedBlocksOwner(owner []byte) error
	SetValidatorMinimumStakeOwner(owner []byte) error
	SetValidatorUnstakingBlocksOwner(owner []byte) error
	SetValidatorMinimumPauseBlocksOwner(owner []byte) error
	SetValidatorMaxPausedBlocksOwner(owner []byte) error
	SetValidatorMaximumMissedBlocksOwner(owner []byte) error
	SetProposerPercentageOfFeesOwner(owner []byte) error
	SetMaxEvidenceAgeInBlocksOwner(owner []byte) error
	SetMissedBlocksBurnPercentageOwner(owner []byte) error
	SetDoubleSignBurnPercentageOwner(owner []byte) error
	SetServiceNodesPerSessionOwner(owner []byte) error
}

type PersistenceReadContext interface {
	GetHeight() (int64, error)

	// Block Queries
	GetLatestBlockHeight() (uint64, error)
	GetBlockHash(height int64) ([]byte, error)
	GetBlocksPerSession() (int, error)

	// Indexer Queries
	TransactionExists(transactionHash string) (bool, error)

	// Pool Queries
	GetPoolAmount(name string, height int64) (amount string, err error)

	// Account Queries
	GetAccountAmount(address []byte, height int64) (string, error)

	// App Queries
	GetAppExists(address []byte, height int64) (exists bool, err error)
	GetAppStakeAmount(height int64, address []byte) (string, error)
	GetAppsReadyToUnstake(height int64, status int) (apps []*types.UnstakingActor, err error)
	GetAppStatus(address []byte, height int64) (status int, err error)
	GetAppPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetAppOutputAddress(operator []byte, height int64) (output []byte, err error)

	// ServiceNode Queries
	GetServiceNodeExists(address []byte, height int64) (exists bool, err error)
	GetServiceNodeStakeAmount(height int64, address []byte) (string, error)
	GetServiceNodesReadyToUnstake(height int64, status int) (serviceNodes []*types.UnstakingActor, err error)
	GetServiceNodeStatus(address []byte, height int64) (status int, err error)
	GetServiceNodePauseHeightIfExists(address []byte, height int64) (int64, error)
	GetServiceNodeOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetServiceNodeCount(chain string, height int64) (int, error)
	GetServiceNodesPerSessionAt(height int64) (int, error)

	// Fisherman Queries
	GetFishermanExists(address []byte, height int64) (exists bool, err error)
	GetFishermanStakeAmount(height int64, address []byte) (string, error)
	GetFishermenReadyToUnstake(height int64, status int) (fishermen []*types.UnstakingActor, err error)
	GetFishermanStatus(address []byte, height int64) (status int, err error)
	GetFishermanPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetFishermanOutputAddress(operator []byte, height int64) (output []byte, err error)

	// Validator Queries
	GetValidatorExists(address []byte, height int64) (exists bool, err error)
	GetValidatorStakeAmount(height int64, address []byte) (string, error)
	GetValidatorsReadyToUnstake(height int64, status int) (validators []*types.UnstakingActor, err error)
	GetValidatorStatus(address []byte, height int64) (status int, err error)
	GetValidatorPauseHeightIfExists(address []byte, height int64) (int64, error)
	GetValidatorOutputAddress(operator []byte, height int64) (output []byte, err error)
	GetValidatorMissedBlocks(address []byte, height int64) (int, error)

	/* TODO(olshansky): review/revisit this in more details */

	// Param Queries
	GetParamAppMinimumStake() (string, error)
	GetMaxAppChains() (int, error)
	GetBaselineAppStakeRate() (int, error)
	GetStabilityAdjustment() (int, error)
	GetAppUnstakingBlocks() (int, error)
	GetAppMinimumPauseBlocks() (int, error)
	GetAppMaxPausedBlocks() (int, error)

	GetParamServiceNodeMinimumStake() (string, error)
	GetServiceNodeMaxChains() (int, error)
	GetServiceNodeUnstakingBlocks() (int, error)
	GetServiceNodeMinimumPauseBlocks() (int, error)
	GetServiceNodeMaxPausedBlocks() (int, error)
	GetServiceNodesPerSession() (int, error)

	GetParamFishermanMinimumStake() (string, error)
	GetFishermanMaxChains() (int, error)
	GetFishermanUnstakingBlocks() (int, error)
	GetFishermanMinimumPauseBlocks() (int, error)
	GetFishermanMaxPausedBlocks() (int, error)

	GetParamValidatorMinimumStake() (string, error)
	GetValidatorUnstakingBlocks() (int, error)
	GetValidatorMinimumPauseBlocks() (int, error)
	GetValidatorMaxPausedBlocks() (int, error)
	GetValidatorMaximumMissedBlocks() (int, error)
	GetProposerPercentageOfFees() (int, error)
	GetMaxEvidenceAgeInBlocks() (int, error)
	GetMissedBlocksBurnPercentage() (int, error)
	GetDoubleSignBurnPercentage() (int, error)

	GetMessageDoubleSignFee() (string, error)
	GetMessageSendFee() (string, error)
	GetMessageStakeFishermanFee() (string, error)
	GetMessageEditStakeFishermanFee() (string, error)
	GetMessageUnstakeFishermanFee() (string, error)
	GetMessagePauseFishermanFee() (string, error)
	GetMessageUnpauseFishermanFee() (string, error)
	GetMessageFishermanPauseServiceNodeFee() (string, error)
	GetMessageTestScoreFee() (string, error)
	GetMessageProveTestScoreFee() (string, error)
	GetMessageStakeAppFee() (string, error)
	GetMessageEditStakeAppFee() (string, error)
	GetMessageUnstakeAppFee() (string, error)
	GetMessagePauseAppFee() (string, error)
	GetMessageUnpauseAppFee() (string, error)
	GetMessageStakeValidatorFee() (string, error)
	GetMessageEditStakeValidatorFee() (string, error)
	GetMessageUnstakeValidatorFee() (string, error)
	GetMessagePauseValidatorFee() (string, error)
	GetMessageUnpauseValidatorFee() (string, error)
	GetMessageStakeServiceNodeFee() (string, error)
	GetMessageEditStakeServiceNodeFee() (string, error)
	GetMessageUnstakeServiceNodeFee() (string, error)
	GetMessagePauseServiceNodeFee() (string, error)
	GetMessageUnpauseServiceNodeFee() (string, error)
	GetMessageChangeParameterFee() (string, error)

	// ACL Queries
	GetAclOwner() ([]byte, error)
	GetBlocksPerSessionOwner() ([]byte, error)
	GetMaxAppChainsOwner() ([]byte, error)
	GetAppMinimumStakeOwner() ([]byte, error)
	GetBaselineAppOwner() ([]byte, error)
	GetStakingAdjustmentOwner() ([]byte, error)
	GetAppUnstakingBlocksOwner() ([]byte, error)
	GetAppMinimumPauseBlocksOwner() ([]byte, error)
	GetAppMaxPausedBlocksOwner() ([]byte, error)
	GetParamServiceNodeMinimumStakeOwner() ([]byte, error)
	GetServiceNodeMaxChainsOwner() ([]byte, error)
	GetServiceNodeUnstakingBlocksOwner() ([]byte, error)
	GetServiceNodeMinimumPauseBlocksOwner() ([]byte, error)
	GetServiceNodeMaxPausedBlocksOwner() ([]byte, error)
	GetFishermanMinimumStakeOwner() ([]byte, error)
	GetMaxFishermanChainsOwner() ([]byte, error)
	GetFishermanUnstakingBlocksOwner() ([]byte, error)
	GetFishermanMinimumPauseBlocksOwner() ([]byte, error)
	GetFishermanMaxPausedBlocksOwner() ([]byte, error)
	GetValidatorMinimumStakeOwner() ([]byte, error)
	GetValidatorUnstakingBlocksOwner() ([]byte, error)
	GetValidatorMinimumPauseBlocksOwner() ([]byte, error)
	GetValidatorMaxPausedBlocksOwner() ([]byte, error)
	GetValidatorMaximumMissedBlocksOwner() ([]byte, error)
	GetProposerPercentageOfFeesOwner() ([]byte, error)
	GetMaxEvidenceAgeInBlocksOwner() ([]byte, error)
	GetMissedBlocksBurnPercentageOwner() ([]byte, error)
	GetDoubleSignBurnPercentageOwner() ([]byte, error)
	GetServiceNodesPerSessionOwner() ([]byte, error)
	GetMessageDoubleSignFeeOwner() ([]byte, error)
	GetMessageSendFeeOwner() ([]byte, error)
	GetMessageStakeFishermanFeeOwner() ([]byte, error)
	GetMessageEditStakeFishermanFeeOwner() ([]byte, error)
	GetMessageUnstakeFishermanFeeOwner() ([]byte, error)
	GetMessagePauseFishermanFeeOwner() ([]byte, error)
	GetMessageUnpauseFishermanFeeOwner() ([]byte, error)
	GetMessageFishermanPauseServiceNodeFeeOwner() ([]byte, error)
	GetMessageTestScoreFeeOwner() ([]byte, error)
	GetMessageProveTestScoreFeeOwner() ([]byte, error)
	GetMessageStakeAppFeeOwner() ([]byte, error)
	GetMessageEditStakeAppFeeOwner() ([]byte, error)
	GetMessageUnstakeAppFeeOwner() ([]byte, error)
	GetMessagePauseAppFeeOwner() ([]byte, error)
	GetMessageUnpauseAppFeeOwner() ([]byte, error)
	GetMessageStakeValidatorFeeOwner() ([]byte, error)
	GetMessageEditStakeValidatorFeeOwner() ([]byte, error)
	GetMessageUnstakeValidatorFeeOwner() ([]byte, error)
	GetMessagePauseValidatorFeeOwner() ([]byte, error)
	GetMessageUnpauseValidatorFeeOwner() ([]byte, error)
	GetMessageStakeServiceNodeFeeOwner() ([]byte, error)
	GetMessageEditStakeServiceNodeFeeOwner() ([]byte, error)
	GetMessageUnstakeServiceNodeFeeOwner() ([]byte, error)
	GetMessagePauseServiceNodeFeeOwner() ([]byte, error)
	GetMessageUnpauseServiceNodeFeeOwner() ([]byte, error)
	GetMessageChangeParameterFeeOwner() ([]byte, error)
}
