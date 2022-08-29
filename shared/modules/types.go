package modules

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file contains the minimum shared structures (GenesisState) and the many shared interfaces the modules implement
// the main purpose of this structure is to ensure the ownership of the

type GenesisState struct {
	PersistenceGenesisState PersistenceGenesisState
	ConsensusGenesisState   ConsensusGenesisState
}

type BaseConfig struct {
	RootDirectory string `json:"root_directory"`
	PrivateKey    string `json:"private_key"` // TODO (team) better architecture for key management (keybase, keyfiles, etc.)
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
	IsEmptyConnType() bool // TODO (team) make enum
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
	GetActorTyp() ActorType // TODO (research) this method has to be implemented manually which is a pain
}

type ActorType interface {
	String() string
}

type UnstakingActorI interface {
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

// TODO (Team) move to use proto string() and deprecate #147
const (
	BlocksPerSessionParamName = "blocks_per_session"

	AppMinimumStakeParamName       = "app_minimum_stake"
	AppMaxChainsParamName          = "app_max_chains"
	AppBaselineStakeRateParamName  = "app_baseline_stake_rate"
	AppStakingAdjustmentParamName  = "app_staking_adjustment"
	AppUnstakingBlocksParamName    = "app_unstaking_blocks"
	AppMinimumPauseBlocksParamName = "app_minimum_pause_blocks"
	AppMaxPauseBlocksParamName     = "app_max_pause_blocks"

	ServiceNodeMinimumStakeParamName       = "service_node_minimum_stake"
	ServiceNodeMaxChainsParamName          = "service_node_max_chains"
	ServiceNodeUnstakingBlocksParamName    = "service_node_unstaking_blocks"
	ServiceNodeMinimumPauseBlocksParamName = "service_node_minimum_pause_blocks"
	ServiceNodeMaxPauseBlocksParamName     = "service_node_max_pause_blocks"
	ServiceNodesPerSessionParamName        = "service_nodes_per_session"

	FishermanMinimumStakeParamName       = "fisherman_minimum_stake"
	FishermanMaxChainsParamName          = "fisherman_max_chains"
	FishermanUnstakingBlocksParamName    = "fisherman_unstaking_blocks"
	FishermanMinimumPauseBlocksParamName = "fisherman_minimum_pause_blocks"
	FishermanMaxPauseBlocksParamName     = "fisherman_max_pause_blocks"

	ValidatorMinimumStakeParamName        = "validator_minimum_stake"
	ValidatorUnstakingBlocksParamName     = "validator_unstaking_blocks"
	ValidatorMinimumPauseBlocksParamName  = "validator_minimum_pause_blocks"
	ValidatorMaxPausedBlocksParamName     = "validator_max_pause_blocks"
	ValidatorMaximumMissedBlocksParamName = "validator_maximum_missed_blocks"

	ValidatorMaxEvidenceAgeInBlocksParamName = "validator_max_evidence_age_in_blocks"
	ProposerPercentageOfFeesParamName        = "proposer_percentage_of_fees"
	MissedBlocksBurnPercentageParamName      = "missed_blocks_burn_percentage"
	DoubleSignBurnPercentageParamName        = "double_sign_burn_percentage"

	MessageDoubleSignFee                = "message_double_sign_fee"
	MessageSendFee                      = "message_send_fee"
	MessageStakeFishermanFee            = "message_stake_fisherman_fee"
	MessageEditStakeFishermanFee        = "message_edit_stake_fisherman_fee"
	MessageUnstakeFishermanFee          = "message_unstake_fisherman_fee"
	MessagePauseFishermanFee            = "message_pause_fisherman_fee"
	MessageUnpauseFishermanFee          = "message_unpause_fisherman_fee"
	MessageFishermanPauseServiceNodeFee = "message_fisherman_pause_service_node_fee"
	MessageTestScoreFee                 = "message_test_score_fee"
	MessageProveTestScoreFee            = "message_prove_test_score_fee"
	MessageStakeAppFee                  = "message_stake_app_fee"
	MessageEditStakeAppFee              = "message_edit_stake_app_fee"
	MessageUnstakeAppFee                = "message_unstake_app_fee"
	MessagePauseAppFee                  = "message_pause_app_fee"
	MessageUnpauseAppFee                = "message_unpause_app_fee"
	MessageStakeValidatorFee            = "message_stake_validator_fee"
	MessageEditStakeValidatorFee        = "message_edit_stake_validator_fee"
	MessageUnstakeValidatorFee          = "message_unstake_validator_fee"
	MessagePauseValidatorFee            = "message_pause_validator_fee"
	MessageUnpauseValidatorFee          = "message_unpause_validator_fee"
	MessageStakeServiceNodeFee          = "message_stake_service_node_fee"
	MessageEditStakeServiceNodeFee      = "message_edit_stake_service_node_fee"
	MessageUnstakeServiceNodeFee        = "message_unstake_service_node_fee"
	MessagePauseServiceNodeFee          = "message_pause_service_node_fee"
	MessageUnpauseServiceNodeFee        = "message_unpause_service_node_fee"
	MessageChangeParameterFee           = "message_change_parameter_fee"

	AclOwner                                 = "acl_owner"
	BlocksPerSessionOwner                    = "blocks_per_session_owner"
	AppMinimumStakeOwner                     = "app_minimum_stake_owner"
	AppMaxChainsOwner                        = "app_max_chains_owner"
	AppBaselineStakeRateOwner                = "app_baseline_stake_rate_owner"
	AppStakingAdjustmentOwner                = "app_staking_adjustment_owner"
	AppUnstakingBlocksOwner                  = "app_unstaking_blocks_owner"
	AppMinimumPauseBlocksOwner               = "app_minimum_pause_blocks_owner"
	AppMaxPausedBlocksOwner                  = "app_max_paused_blocks_owner"
	ServiceNodeMinimumStakeOwner             = "service_node_minimum_stake_owner"
	ServiceNodeMaxChainsOwner                = "service_node_max_chains_owner"
	ServiceNodeUnstakingBlocksOwner          = "service_node_unstaking_blocks_owner"
	ServiceNodeMinimumPauseBlocksOwner       = "service_node_minimum_pause_blocks_owner"
	ServiceNodeMaxPausedBlocksOwner          = "service_node_max_paused_blocks_owner"
	ServiceNodesPerSessionOwner              = "service_nodes_per_session_owner"
	FishermanMinimumStakeOwner               = "fisherman_minimum_stake_owner"
	FishermanMaxChainsOwner                  = "fisherman_max_chains_owner"
	FishermanUnstakingBlocksOwner            = "fisherman_unstaking_blocks_owner"
	FishermanMinimumPauseBlocksOwner         = "fisherman_minimum_pause_blocks_owner"
	FishermanMaxPausedBlocksOwner            = "fisherman_max_paused_blocks_owner"
	ValidatorMinimumStakeOwner               = "validator_minimum_stake_owner"
	ValidatorUnstakingBlocksOwner            = "validator_unstaking_blocks_owner"
	ValidatorMinimumPauseBlocksOwner         = "validator_minimum_pause_blocks_owner"
	ValidatorMaxPausedBlocksOwner            = "validator_max_paused_blocks_owner"
	ValidatorMaximumMissedBlocksOwner        = "validator_maximum_missed_blocks_owner"
	ValidatorMaxEvidenceAgeInBlocksOwner     = "validator_max_evidence_age_in_blocks_owner"
	ProposerPercentageOfFeesOwner            = "proposer_percentage_of_fees_owner"
	MissedBlocksBurnPercentageOwner          = "missed_blocks_burn_percentage_owner"
	DoubleSignBurnPercentageOwner            = "double_sign_burn_percentage_owner"
	MessageDoubleSignFeeOwner                = "message_double_sign_fee_owner"
	MessageSendFeeOwner                      = "message_send_fee_owner"
	MessageStakeFishermanFeeOwner            = "message_stake_fisherman_fee_owner"
	MessageEditStakeFishermanFeeOwner        = "message_edit_stake_fisherman_fee_owner"
	MessageUnstakeFishermanFeeOwner          = "message_unstake_fisherman_fee_owner"
	MessagePauseFishermanFeeOwner            = "message_pause_fisherman_fee_owner"
	MessageUnpauseFishermanFeeOwner          = "message_unpause_fisherman_fee_owner"
	MessageFishermanPauseServiceNodeFeeOwner = "message_fisherman_pause_service_node_fee_owner"
	MessageTestScoreFeeOwner                 = "message_test_score_fee_owner"
	MessageProveTestScoreFeeOwner            = "message_prove_test_score_fee_owner"
	MessageStakeAppFeeOwner                  = "message_stake_app_fee_owner"
	MessageEditStakeAppFeeOwner              = "message_edit_stake_app_fee_owner"
	MessageUnstakeAppFeeOwner                = "message_unstake_app_fee_owner"
	MessagePauseAppFeeOwner                  = "message_pause_app_fee_owner"
	MessageUnpauseAppFeeOwner                = "message_unpause_app_fee_owner"
	MessageStakeValidatorFeeOwner            = "message_stake_validator_fee_owner"
	MessageEditStakeValidatorFeeOwner        = "message_edit_stake_validator_fee_owner"
	MessageUnstakeValidatorFeeOwner          = "message_unstake_validator_fee_owner"
	MessagePauseValidatorFeeOwner            = "message_pause_validator_fee_owner"
	MessageUnpauseValidatorFeeOwner          = "message_unpause_validator_fee_owner"
	MessageStakeServiceNodeFeeOwner          = "message_stake_service_node_fee_owner"
	MessageEditStakeServiceNodeFeeOwner      = "message_edit_stake_service_node_fee_owner"
	MessageUnstakeServiceNodeFeeOwner        = "message_unstake_service_node_fee_owner"
	MessagePauseServiceNodeFeeOwner          = "message_pause_service_node_fee_owner"
	MessageUnpauseServiceNodeFeeOwner        = "message_unpause_service_node_fee_owner"
	MessageChangeParameterFeeOwner           = "message_change_parameter_fee_owner"
)
