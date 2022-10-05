package modules

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file contains the minimum shared structures (GenesisState) and the many shared interfaces the modules implement
// the main purpose of this structure is to ensure the ownership of the

type GenesisState struct {
	PersistenceGenesisState PersistenceGenesisState `json:"persistence_genesis_state"`
	ConsensusGenesisState   ConsensusGenesisState   `json:"consensus_genesis_state"`
}

type BaseConfig struct {
	RootDirectory string `json:"root_directory"`
	PrivateKey    string `json:"private_key"` // TODO (#150) better architecture for key management (keybase, keyfiles, etc.)
}

type Config struct {
	Base        *BaseConfig       `json:"base"`
	Consensus   ConsensusConfig   `json:"consensus"`
	Utility     UtilityConfig     `json:"utility"`
	Persistence PersistenceConfig `json:"persistence"`
	P2P         P2PConfig         `json:"p2p"`
	Telemetry   TelemetryConfig   `json:"telemetry"`
}

type ConsensusConfig interface {
	GetMaxMempoolBytes() uint64
	GetPaceMakerConfig() PacemakerConfig
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
}

type P2PConfig interface {
	GetConsensusPort() uint32
	GetUseRainTree() bool
	IsEmptyConnType() bool // TODO : make enum
}

type TelemetryConfig interface {
	GetEnabled() bool
	GetAddress() string
	GetEndpoint() string
}

type UtilityConfig interface{}

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

var _ IConfig = PacemakerConfig(nil)
var _ IConfig = PersistenceConfig(nil)
var _ IConfig = P2PConfig(nil)
var _ IConfig = TelemetryConfig(nil)
var _ IConfig = UtilityConfig(nil)

var _ IGenesis = PersistenceGenesisState(nil)
var _ IGenesis = ConsensusGenesisState(nil)

// TODO(#235): Remove these interfaces once the runtime config approach is implemented.
type IConfig interface{}
type IGenesis interface{}
