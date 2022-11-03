package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/types_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file contains the minimum shared structures (GenesisState) and the many shared interfaces the modules implement
// the main purpose of this structure is to ensure the ownership of the

type GenesisState interface {
	GetPersistenceGenesisState() PersistenceGenesisState
	GetConsensusGenesisState() ConsensusGenesisState
}

type Config interface {
	GetBaseConfig() BaseConfig
	GetConsensusConfig() ConsensusConfig
	GetUtilityConfig() UtilityConfig
	GetPersistenceConfig() PersistenceConfig
	GetP2PConfig() P2PConfig
	GetTelemetryConfig() TelemetryConfig
	GetLoggerConfig() LoggerConfig
}

type BaseConfig interface {
	GetRootDirectory() string
	GetPrivateKey() string // TODO (#150) better architecture for key management (keybase, keyfiles, etc.)
}

type ConsensusConfig interface {
	GetMaxMempoolBytes() uint64
	GetPrivateKey() string // TODO (#150) better architecture for key management (keybase, keyfiles, etc.)
}

type PacemakerConfig interface {
	SetTimeoutMsec(uint64)
	GetTimeoutMsec() uint64
	GetManual() bool
	GetDebugTimeBetweenStepsMsec() uint64
}

type PersistenceConfig interface {
	GetPostgresUrl() string
	GetNodeSchema() string
	GetBlockStorePath() string
	GetTxIndexerPath() string
}

type P2PConfig interface {
	GetPrivateKey() string
	GetConsensusPort() uint32
	GetUseRainTree() bool
	GetIsEmptyConnectionType() bool // TODO : make enum
}

type TelemetryConfig interface {
	GetEnabled() bool
	GetAddress() string
	GetEndpoint() string
}

type LoggerConfig interface {
	// We have protobuf enums for the following values, but they are represented as
	// `strings` to avoid circular dependencies.
	GetLevel() string
	GetFormat() string
}

type UtilityConfig interface {
	GetMaxMempoolTransactionBytes() uint64
	GetMaxMempoolTransactions() uint32
}

type RPCConfig interface {
	GetEnabled() bool
	GetPort() string
	GetTimeout() uint64
}

type PersistenceGenesisState interface {
	GetAccs() []Account
	GetAccPools() []Account
	GetApps() []Actor
	GetVals() []Actor
	GetFish() []Actor
	GetNodes() []Actor
	GetParameters() Params
}

type ConsensusGenesisState interface {
	GetGenesisTime() *timestamppb.Timestamp
	GetChainId() string
	GetMaxBlockBytes() uint64
	GetVals() []Actor
}

type Account interface {
	GetAddress() string
	GetAmount() string
}

type Actor interface {
	GetAddress() string
	GetPublicKey() string
	GetChains() []string
	GetGenericParam() string
	GetStakedAmount() string
	GetPausedHeight() int64
	GetUnstakingHeight() int64
	GetOutput() string
	GetActorTyp() ActorType // RESEARCH: this method has to be implemented manually which is a pain
}

type ActorType interface {
	String() string
}

type IUnstakingActor interface {
	GetAddress() []byte
	SetAddress(address string)
	GetStakeAmount() string
	SetStakeAmount(address string)
	GetOutputAddress() []byte
	SetOutputAddress(address string)
}

type Params interface {
	GetAppMinimumStake() string
	GetAppMaxChains() int32
	GetAppBaselineStakeRate() int32
	GetAppStakingAdjustment() int32
	GetAppUnstakingBlocks() int32
	GetAppMinimumPauseBlocks() int32
	GetAppMaxPauseBlocks() int32
	GetBlocksPerSession() int32

	GetServiceNodeMinimumStake() string
	GetServiceNodeMaxChains() int32
	GetServiceNodeUnstakingBlocks() int32
	GetServiceNodeMinimumPauseBlocks() int32
	GetServiceNodeMaxPauseBlocks() int32
	GetServiceNodesPerSession() int32

	GetFishermanMinimumStake() string
	GetFishermanMaxChains() int32
	GetFishermanUnstakingBlocks() int32
	GetFishermanMinimumPauseBlocks() int32
	GetFishermanMaxPauseBlocks() int32

	GetValidatorMinimumStake() string
	GetValidatorUnstakingBlocks() int32
	GetValidatorMinimumPauseBlocks() int32
	GetValidatorMaxPauseBlocks() int32
	GetValidatorMaximumMissedBlocks() int32
	GetProposerPercentageOfFees() int32
	GetValidatorMaxEvidenceAgeInBlocks() int32
	GetMissedBlocksBurnPercentage() int32
	GetDoubleSignBurnPercentage() int32

	GetMessageDoubleSignFee() string
	GetMessageSendFee() string
	GetMessageStakeFishermanFee() string
	GetMessageEditStakeFishermanFee() string
	GetMessageUnstakeFishermanFee() string
	GetMessagePauseFishermanFee() string
	GetMessageUnpauseFishermanFee() string
	GetMessageFishermanPauseServiceNodeFee() string
	GetMessageTestScoreFee() string
	GetMessageProveTestScoreFee() string
	GetMessageStakeAppFee() string
	GetMessageEditStakeAppFee() string
	GetMessageUnstakeAppFee() string
	GetMessagePauseAppFee() string
	GetMessageUnpauseAppFee() string
	GetMessageStakeValidatorFee() string
	GetMessageEditStakeValidatorFee() string
	GetMessageUnstakeValidatorFee() string
	GetMessagePauseValidatorFee() string
	GetMessageUnpauseValidatorFee() string
	GetMessageStakeServiceNodeFee() string
	GetMessageEditStakeServiceNodeFee() string
	GetMessageUnstakeServiceNodeFee() string
	GetMessagePauseServiceNodeFee() string
	GetMessageUnpauseServiceNodeFee() string
	GetMessageChangeParameterFee() string

	// ACL Queries
	GetAclOwner() string
	GetBlocksPerSessionOwner() string
	GetAppMaxChainsOwner() string
	GetAppMinimumStakeOwner() string
	GetAppBaselineStakeRateOwner() string
	GetAppStakingAdjustmentOwner() string
	GetAppUnstakingBlocksOwner() string
	GetAppMinimumPauseBlocksOwner() string
	GetAppMaxPausedBlocksOwner() string
	GetServiceNodeMinimumStakeOwner() string
	GetServiceNodeMaxChainsOwner() string
	GetServiceNodeUnstakingBlocksOwner() string
	GetServiceNodeMinimumPauseBlocksOwner() string
	GetServiceNodeMaxPausedBlocksOwner() string
	GetFishermanMinimumStakeOwner() string
	GetFishermanMaxChainsOwner() string
	GetFishermanUnstakingBlocksOwner() string
	GetFishermanMinimumPauseBlocksOwner() string
	GetFishermanMaxPausedBlocksOwner() string
	GetValidatorMinimumStakeOwner() string
	GetValidatorUnstakingBlocksOwner() string
	GetValidatorMinimumPauseBlocksOwner() string
	GetValidatorMaxPausedBlocksOwner() string
	GetValidatorMaximumMissedBlocksOwner() string
	GetProposerPercentageOfFeesOwner() string
	GetValidatorMaxEvidenceAgeInBlocksOwner() string
	GetMissedBlocksBurnPercentageOwner() string
	GetDoubleSignBurnPercentageOwner() string
	GetServiceNodesPerSessionOwner() string
	GetMessageDoubleSignFeeOwner() string
	GetMessageSendFeeOwner() string
	GetMessageStakeFishermanFeeOwner() string
	GetMessageEditStakeFishermanFeeOwner() string
	GetMessageUnstakeFishermanFeeOwner() string
	GetMessagePauseFishermanFeeOwner() string
	GetMessageUnpauseFishermanFeeOwner() string
	GetMessageFishermanPauseServiceNodeFeeOwner() string
	GetMessageTestScoreFeeOwner() string
	GetMessageProveTestScoreFeeOwner() string
	GetMessageStakeAppFeeOwner() string
	GetMessageEditStakeAppFeeOwner() string
	GetMessageUnstakeAppFeeOwner() string
	GetMessagePauseAppFeeOwner() string
	GetMessageUnpauseAppFeeOwner() string
	GetMessageStakeValidatorFeeOwner() string
	GetMessageEditStakeValidatorFeeOwner() string
	GetMessageUnstakeValidatorFeeOwner() string
	GetMessagePauseValidatorFeeOwner() string
	GetMessageUnpauseValidatorFeeOwner() string
	GetMessageStakeServiceNodeFeeOwner() string
	GetMessageEditStakeServiceNodeFeeOwner() string
	GetMessageUnstakeServiceNodeFeeOwner() string
	GetMessagePauseServiceNodeFeeOwner() string
	GetMessageUnpauseServiceNodeFeeOwner() string
	GetMessageChangeParameterFeeOwner() string
}

// The result of executing a transaction against the blockchain state so that it is included in the block
type TxResult interface {
	GetTx() []byte                        // the transaction object primitive
	GetHeight() int64                     // the height at which the tx was applied
	GetIndex() int32                      // the transaction's index within the block (i.e. ordered by when the proposer received it in the mempool)
	GetResultCode() int32                 // 0 is no error, otherwise corresponds to error object code; // IMPROVE: Add a specific type fot he result code
	GetError() string                     // can be empty; IMPROVE: Add a specific type fot he error code
	GetSignerAddr() string                // get the address of who signed (i.e. sent) the transaction
	GetRecipientAddr() string             // get the address of who received the transaction; may be empty
	GetMessageType() string               // corresponds to type of message (validator-stake, app-unjail, node-stake, etc) // IMPROVE: Add an enum for message types
	Hash() ([]byte, error)                // the hash of the tx bytes
	HashFromBytes([]byte) ([]byte, error) // same operation as `Hash`, but avoid re-serializing the tx
	Bytes() ([]byte, error)               // returns the serialized transaction bytes
	FromBytes([]byte) (TxResult, error)   // returns the deserialized transaction result
}
