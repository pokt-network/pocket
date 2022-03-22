package modules

import (
	"github.com/pokt-network/pocket/shared/types"
	"github.com/syndtr/goleveldb/leveldb/memdb"
)

type PersistenceModule interface {
	Module

	NewContext(height int64) (PersistenceContext, error)
	GetCommitDB() *memdb.DB
}

type PersistenceContext interface {
	GetLatestBlockHeight() (uint64, error)
	GetBlockHash(height int64) ([]byte, error)

	// Context Operations
	NewSavePoint([]byte) error
	RollbackToSavePoint([]byte) error
	AppHash() ([]byte, error)
	Reset() error
	Commit() error
	Release()
	GetHeight() (int64, error)

	// Indexer
	TransactionExists(transactionHash string) bool

	// Account
	AddPoolAmount(name string, amount string) error
	SubtractPoolAmount(name string, amount string) error
	SetPoolAmount(name string, amount string) error
	InsertPool(name string, address []byte, amount string) error
	GetPoolAmount(name string) (amount string, err error)
	AddAccountAmount(address []byte, amount string) error
	SubtractAccountAmount(address []byte, amount string) error
	GetAccountAmount(address []byte) (string, error)
	SetAccount(address []byte, amount string) error

	// App
	GetAppExists(address []byte) (exists bool, err error)
	InsertApplication(address []byte, publicKey []byte, output []byte, paused bool, status int, maxRelays string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateApplication(address []byte, maxRelaysToAdd string, amountToAdd string, chainsToUpdate []string) error
	DeleteApplication(address []byte) error
	GetAppsReadyToUnstake(Height int64, status int) (apps []*types.UnstakingActor, err error)
	GetAppStatus(address []byte) (status int, err error)
	SetAppUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetAppPauseHeightIfExists(address []byte) (int64, error)
	SetAppsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetAppPauseHeight(address []byte, height int64) error
	GetAppOutputAddress(operator []byte) (output []byte, err error)

	// ServiceNode
	GetServiceNodeExists(address []byte) (exists bool, err error)
	InsertServiceNode(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateServiceNode(address []byte, serviceURL string, amountToAdd string, chains []string) error
	DeleteServiceNode(address []byte) error
	GetServiceNodesReadyToUnstake(Height int64, status int) (ServiceNodes []*types.UnstakingActor, err error)
	GetServiceNodeStatus(address []byte) (status int, err error)
	SetServiceNodeUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetServiceNodePauseHeightIfExists(address []byte) (int64, error)
	SetServiceNodesStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetServiceNodePauseHeight(address []byte, height int64) error
	GetServiceNodesPerSessionAt(height int64) (int, error)
	GetServiceNodeCount(chain string, height int64) (int, error)
	GetServiceNodeOutputAddress(operator []byte) (output []byte, err error)

	// Fisherman
	GetFishermanExists(address []byte) (exists bool, err error)
	InsertFisherman(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, chains []string, pausedHeight int64, unstakingHeight int64) error
	UpdateFisherman(address []byte, serviceURL string, amountToAdd string, chains []string) error
	DeleteFisherman(address []byte) error
	GetFishermanReadyToUnstake(Height int64, status int) (Fishermans []*types.UnstakingActor, err error)
	GetFishermanStatus(address []byte) (status int, err error)
	SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetFishermanPauseHeightIfExists(address []byte) (int64, error)
	SetFishermansStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetFishermanPauseHeight(address []byte, height int64) error
	GetFishermanOutputAddress(operator []byte) (output []byte, err error)

	// Validator
	GetValidatorExists(address []byte) (exists bool, err error)
	InsertValidator(address []byte, publicKey []byte, output []byte, paused bool, status int, serviceURL string, stakedTokens string, pausedHeight int64, unstakingHeight int64) error
	UpdateValidator(address []byte, serviceURL string, amountToAdd string) error
	DeleteValidator(address []byte) error
	GetValidatorsReadyToUnstake(Height int64, status int) (Validators []*types.UnstakingActor, err error)
	GetValidatorStatus(address []byte) (status int, err error)
	SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) error
	GetValidatorPauseHeightIfExists(address []byte) (int64, error)
	SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight int64, status int) error
	SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) error
	SetValidatorMissedBlocks(address []byte, missedBlocks int) error
	GetValidatorMissedBlocks(address []byte) (int, error)
	SetValidatorPauseHeight(address []byte, height int64) error
	SetValidatorStakedTokens(address []byte, tokens string) error
	GetValidatorStakedTokens(address []byte) (tokens string, err error)
	GetValidatorOutputAddress(operator []byte) (output []byte, err error)

	// Params
	InitParams() error

	GetBlocksPerSession() (int, error)

	GetParamAppMinimumStake() (string, error)
	GetMaxAppChains() (int, error)
	GetBaselineAppStakeRate() (int, error)
	GetStakingAdjustment() (int, error)
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

	// Setters
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

	// ACL
	GetAclOwner() ([]byte, error)
	SetAclOwner(owner []byte) error
	SetBlocksPerSessionOwner(owner []byte) error
	GetBlocksPerSessionOwner() ([]byte, error)
	GetMaxAppChainsOwner() ([]byte, error)
	SetMaxAppChainsOwner(owner []byte) error
	GetAppMinimumStakeOwner() ([]byte, error)
	SetAppMinimumStakeOwner(owner []byte) error
	GetBaselineAppOwner() ([]byte, error)
	SetBaselineAppOwner(owner []byte) error
	GetStakingAdjustmentOwner() ([]byte, error)
	SetStakingAdjustmentOwner(owner []byte) error
	GetAppUnstakingBlocksOwner() ([]byte, error)
	SetAppUnstakingBlocksOwner(owner []byte) error
	GetAppMinimumPauseBlocksOwner() ([]byte, error)
	SetAppMinimumPauseBlocksOwner(owner []byte) error
	GetAppMaxPausedBlocksOwner() ([]byte, error)
	SetAppMaxPausedBlocksOwner(owner []byte) error
	GetParamServiceNodeMinimumStakeOwner() ([]byte, error)
	SetParamServiceNodeMinimumStakeOwner(owner []byte) error
	GetServiceNodeMaxChainsOwner() ([]byte, error)
	SetMaxServiceNodeChainsOwner(owner []byte) error
	GetServiceNodeUnstakingBlocksOwner() ([]byte, error)
	SetServiceNodeUnstakingBlocksOwner(owner []byte) error
	GetServiceNodeMinimumPauseBlocksOwner() ([]byte, error)
	SetServiceNodeMinimumPauseBlocksOwner(owner []byte) error
	GetServiceNodeMaxPausedBlocksOwner() ([]byte, error)
	SetServiceNodeMaxPausedBlocksOwner(owner []byte) error
	GetFishermanMinimumStakeOwner() ([]byte, error)
	SetFishermanMinimumStakeOwner(owner []byte) error
	GetMaxFishermanChainsOwner() ([]byte, error)
	SetMaxFishermanChainsOwner(owner []byte) error
	GetFishermanUnstakingBlocksOwner() ([]byte, error)
	SetFishermanUnstakingBlocksOwner(owner []byte) error
	GetFishermanMinimumPauseBlocksOwner() ([]byte, error)
	SetFishermanMinimumPauseBlocksOwner(owner []byte) error
	GetFishermanMaxPausedBlocksOwner() ([]byte, error)
	SetFishermanMaxPausedBlocksOwner(owner []byte) error
	GetParamValidatorMinimumStakeOwner() ([]byte, error)
	SetParamValidatorMinimumStakeOwner(owner []byte) error
	GetValidatorUnstakingBlocksOwner() ([]byte, error)
	SetValidatorUnstakingBlocksOwner(owner []byte) error
	GetValidatorMinimumPauseBlocksOwner() ([]byte, error)
	SetValidatorMinimumPauseBlocksOwner(owner []byte) error
	GetValidatorMaxPausedBlocksOwner() ([]byte, error)
	SetValidatorMaxPausedBlocksOwner(owner []byte) error
	GetValidatorMaximumMissedBlocksOwner() ([]byte, error)
	SetValidatorMaximumMissedBlocksOwner(owner []byte) error
	GetProposerPercentageOfFeesOwner() ([]byte, error)
	SetProposerPercentageOfFeesOwner(owner []byte) error
	GetMaxEvidenceAgeInBlocksOwner() ([]byte, error)
	SetMaxEvidenceAgeInBlocksOwner(owner []byte) error
	GetMissedBlocksBurnPercentageOwner() ([]byte, error)
	SetMissedBlocksBurnPercentageOwner(owner []byte) error
	GetDoubleSignBurnPercentageOwner() ([]byte, error)
	SetDoubleSignBurnPercentageOwner(owner []byte) error
	SetServiceNodesPerSessionOwner(owner []byte) error
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
