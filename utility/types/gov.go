package types

const (
	// Session gov params
	BlocksPerSessionParamName = "blocks_per_session"

	// Application actor gov params
	AppMinimumStakeParamName       = "app_minimum_stake"
	AppMaxChainsParamName          = "app_max_chains"
	AppBaselineStakeRateParamName  = "app_baseline_stake_rate"
	AppStakingAdjustmentParamName  = "app_staking_adjustment"
	AppUnstakingBlocksParamName    = "app_unstaking_blocks"
	AppMinimumPauseBlocksParamName = "app_minimum_pause_blocks"
	AppMaxPauseBlocksParamName     = "app_max_pause_blocks"

	// Servicer actor gov params
	ServiceNodeMinimumStakeParamName       = "service_node_minimum_stake"
	ServiceNodeMaxChainsParamName          = "service_node_max_chains"
	ServiceNodeUnstakingBlocksParamName    = "service_node_unstaking_blocks"
	ServiceNodeMinimumPauseBlocksParamName = "service_node_minimum_pause_blocks"
	ServiceNodeMaxPauseBlocksParamName     = "service_node_max_pause_blocks"
	ServiceNodesPerSessionParamName        = "service_nodes_per_session"

	// Fisherman actor gov params
	FishermanMinimumStakeParamName       = "fisherman_minimum_stake"
	FishermanMaxChainsParamName          = "fisherman_max_chains"
	FishermanUnstakingBlocksParamName    = "fisherman_unstaking_blocks"
	FishermanMinimumPauseBlocksParamName = "fisherman_minimum_pause_blocks"
	FishermanMaxPauseBlocksParamName     = "fisherman_max_pause_blocks"

	// Validator actor gov params
	ValidatorMinimumStakeParamName        = "validator_minimum_stake"
	ValidatorUnstakingBlocksParamName     = "validator_unstaking_blocks"
	ValidatorMinimumPauseBlocksParamName  = "validator_minimum_pause_blocks"
	ValidatorMaxPausedBlocksParamName     = "validator_max_pause_blocks"
	ValidatorMaximumMissedBlocksParamName = "validator_maximum_missed_blocks"

	// Validator (complex) actor gov params
	ValidatorMaxEvidenceAgeInBlocksParamName = "validator_max_evidence_age_in_blocks"
	ProposerPercentageOfFeesParamName        = "proposer_percentage_of_fees"
	MissedBlocksBurnPercentageParamName      = "missed_blocks_burn_percentage"
	DoubleSignBurnPercentageParamName        = "double_sign_burn_percentage"

	// Pocket specific message gov params
	MessageStakeFishermanFee            = "message_stake_fisherman_fee"
	MessageEditStakeFishermanFee        = "message_edit_stake_fisherman_fee"
	MessageUnstakeFishermanFee          = "message_unstake_fisherman_fee"
	MessagePauseFishermanFee            = "message_pause_fisherman_fee"
	MessageUnpauseFishermanFee          = "message_unpause_fisherman_fee"
	MessageFishermanPauseServiceNodeFee = "message_fisherman_pause_service_node_fee"
	MessageTestScoreFee                 = "message_test_score_fee"
	MessageProveTestScoreFee            = "message_prove_test_score_fee"

	// Proof-of-stake message gov params
	MessageDoubleSignFee   = "message_double_sign_fee"
	MessageSendFee         = "message_send_fee"
	MessageStakeAppFee     = "message_stake_app_fee"
	MessageEditStakeAppFee = "message_edit_stake_app_fee"
	MessageUnstakeAppFee   = "message_unstake_app_fee"
	MessagePauseAppFee     = "message_pause_app_fee"
	MessageUnpauseAppFee   = "message_unpause_app_fee"

	// Validator message gov params
	MessageStakeValidatorFee     = "message_stake_validator_fee"
	MessageEditStakeValidatorFee = "message_edit_stake_validator_fee"
	MessageUnstakeValidatorFee   = "message_unstake_validator_fee"
	MessagePauseValidatorFee     = "message_pause_validator_fee"
	MessageUnpauseValidatorFee   = "message_unpause_validator_fee"

	// Servicer message gov params
	MessageStakeServiceNodeFee     = "message_stake_service_node_fee"
	MessageEditStakeServiceNodeFee = "message_edit_stake_service_node_fee"
	MessageUnstakeServiceNodeFee   = "message_unstake_service_node_fee"
	MessagePauseServiceNodeFee     = "message_pause_service_node_fee"
	MessageUnpauseServiceNodeFee   = "message_unpause_service_node_fee"

	// Parameter / flags gov params
	MessageChangeParameterFee = "message_change_parameter_fee"
)

// TECHDEBT: The parameters below are equivalent to the list above with the suffix `_owner`. There
//           is likely a clean way to better organize this code. This will also involve finding
//           discrepancies between the two lists from missing / duplicate types or owners.
const (
	AclOwner = "acl_owner"

	BlocksPerSessionOwner = "blocks_per_session_owner"

	AppMinimumStakeOwner       = "app_minimum_stake_owner"
	AppMaxChainsOwner          = "app_max_chains_owner"
	AppBaselineStakeRateOwner  = "app_baseline_stake_rate_owner"
	AppStakingAdjustmentOwner  = "app_staking_adjustment_owner"
	AppUnstakingBlocksOwner    = "app_unstaking_blocks_owner"
	AppMinimumPauseBlocksOwner = "app_minimum_pause_blocks_owner"
	AppMaxPausedBlocksOwner    = "app_max_paused_blocks_owner"

	ServiceNodeMinimumStakeOwner       = "service_node_minimum_stake_owner"
	ServiceNodeMaxChainsOwner          = "service_node_max_chains_owner"
	ServiceNodeUnstakingBlocksOwner    = "service_node_unstaking_blocks_owner"
	ServiceNodeMinimumPauseBlocksOwner = "service_node_minimum_pause_blocks_owner"
	ServiceNodeMaxPausedBlocksOwner    = "service_node_max_paused_blocks_owner"
	ServiceNodesPerSessionOwner        = "service_nodes_per_session_owner"

	FishermanMinimumStakeOwner       = "fisherman_minimum_stake_owner"
	FishermanMaxChainsOwner          = "fisherman_max_chains_owner"
	FishermanUnstakingBlocksOwner    = "fisherman_unstaking_blocks_owner"
	FishermanMinimumPauseBlocksOwner = "fisherman_minimum_pause_blocks_owner"
	FishermanMaxPausedBlocksOwner    = "fisherman_max_paused_blocks_owner"

	ValidatorMinimumStakeOwner           = "validator_minimum_stake_owner"
	ValidatorUnstakingBlocksOwner        = "validator_unstaking_blocks_owner"
	ValidatorMinimumPauseBlocksOwner     = "validator_minimum_pause_blocks_owner"
	ValidatorMaxPausedBlocksOwner        = "validator_max_paused_blocks_owner"
	ValidatorMaximumMissedBlocksOwner    = "validator_maximum_missed_blocks_owner"
	ValidatorMaxEvidenceAgeInBlocksOwner = "validator_max_evidence_age_in_blocks_owner"

	ProposerPercentageOfFeesOwner   = "proposer_percentage_of_fees_owner"
	MissedBlocksBurnPercentageOwner = "missed_blocks_burn_percentage_owner"
	DoubleSignBurnPercentageOwner   = "double_sign_burn_percentage_owner"
	MessageDoubleSignFeeOwner       = "message_double_sign_fee_owner"
	MessageSendFeeOwner             = "message_send_fee_owner"

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

	MessageChangeParameterFeeOwner = "message_change_parameter_fee_owner"
)
