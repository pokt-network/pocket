package types

// IMPROVE: Rename `UnstakingBlocks` to `UnbondingPeriod` or `UnstakingBlocksUnbondingPeriod`
// IMPROVE: Create a mapping from ActorType to the gov params that are relevant to that actor type.

const (
	// Session gov params
	BlocksPerSessionParamName = "blocks_per_session"

	// Application actor gov params
	AppMinimumStakeParamName       = "app_minimum_stake"
	AppMaxChainsParamName          = "app_max_chains"
	AppBaselineStakeRateParamName  = "app_baseline_stake_rate"
	AppUnstakingBlocksParamName    = "app_unstaking_blocks"
	AppMinimumPauseBlocksParamName = "app_minimum_pause_blocks"
	AppMaxPauseBlocksParamName     = "app_max_pause_blocks"
	// The constant integer adjustment that the DAO may use to move the stake. The DAO may manually
	// adjust an application's MaxRelays at the time of staking to correct for short-term fluctuations
	// in the price of POKT, which may not be reflected in ParticipationRate
	// When this parameter is set to 0, no adjustment is being made.
	AppStakingAdjustmentParamName = "app_staking_adjustment" // IMPROVE: Document & explain the purpose of this parameter in more detail.

	// Servicer actor gov params
	ServicerMinimumStakeParamName       = "servicer_minimum_stake"
	ServicerMaxChainsParamName          = "servicer_max_chains"
	ServicerUnstakingBlocksParamName    = "servicer_unstaking_blocks"
	ServicerMinimumPauseBlocksParamName = "servicer_minimum_pause_blocks"
	ServicerMaxPauseBlocksParamName     = "servicer_max_pause_blocks"
	ServicersPerSessionParamName        = "servicers_per_session"

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
	MessageStakeFishermanFee         = "message_stake_fisherman_fee"
	MessageEditStakeFishermanFee     = "message_edit_stake_fisherman_fee"
	MessageUnstakeFishermanFee       = "message_unstake_fisherman_fee"
	MessagePauseFishermanFee         = "message_pause_fisherman_fee"
	MessageUnpauseFishermanFee       = "message_unpause_fisherman_fee"
	MessageFishermanPauseServicerFee = "message_fisherman_pause_servicer_fee"
	MessageTestScoreFee              = "message_test_score_fee"
	MessageProveTestScoreFee         = "message_prove_test_score_fee"

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
	MessageStakeServicerFee     = "message_stake_servicer_fee"
	MessageEditStakeServicerFee = "message_edit_stake_servicer_fee"
	MessageUnstakeServicerFee   = "message_unstake_servicer_fee"
	MessagePauseServicerFee     = "message_pause_servicer_fee"
	MessageUnpauseServicerFee   = "message_unpause_servicer_fee"

	// Parameter / flags gov params
	MessageChangeParameterFee = "message_change_parameter_fee"
)

// TECHDEBT: The parameters below are equivalent to the list above with the suffix `_owner`. There
//
//	is likely a clean way to better organize this code. This will also involve finding
//	discrepancies between the two lists from missing / duplicate types or owners.
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

	ServicerMinimumStakeOwner       = "servicer_minimum_stake_owner"
	ServicerMaxChainsOwner          = "servicer_max_chains_owner"
	ServicerUnstakingBlocksOwner    = "servicer_unstaking_blocks_owner"
	ServicerMinimumPauseBlocksOwner = "servicer_minimum_pause_blocks_owner"
	ServicerMaxPausedBlocksOwner    = "servicer_max_paused_blocks_owner"
	ServicersPerSessionOwner        = "servicers_per_session_owner"

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

	MessageStakeFishermanFeeOwner         = "message_stake_fisherman_fee_owner"
	MessageEditStakeFishermanFeeOwner     = "message_edit_stake_fisherman_fee_owner"
	MessageUnstakeFishermanFeeOwner       = "message_unstake_fisherman_fee_owner"
	MessagePauseFishermanFeeOwner         = "message_pause_fisherman_fee_owner"
	MessageUnpauseFishermanFeeOwner       = "message_unpause_fisherman_fee_owner"
	MessageFishermanPauseServicerFeeOwner = "message_fisherman_pause_servicer_fee_owner"
	MessageTestScoreFeeOwner              = "message_test_score_fee_owner"
	MessageProveTestScoreFeeOwner         = "message_prove_test_score_fee_owner"
	MessageStakeAppFeeOwner               = "message_stake_app_fee_owner"
	MessageEditStakeAppFeeOwner           = "message_edit_stake_app_fee_owner"
	MessageUnstakeAppFeeOwner             = "message_unstake_app_fee_owner"
	MessagePauseAppFeeOwner               = "message_pause_app_fee_owner"
	MessageUnpauseAppFeeOwner             = "message_unpause_app_fee_owner"
	MessageStakeValidatorFeeOwner         = "message_stake_validator_fee_owner"
	MessageEditStakeValidatorFeeOwner     = "message_edit_stake_validator_fee_owner"
	MessageUnstakeValidatorFeeOwner       = "message_unstake_validator_fee_owner"
	MessagePauseValidatorFeeOwner         = "message_pause_validator_fee_owner"
	MessageUnpauseValidatorFeeOwner       = "message_unpause_validator_fee_owner"
	MessageStakeServicerFeeOwner          = "message_stake_servicer_fee_owner"
	MessageEditStakeServicerFeeOwner      = "message_edit_stake_servicer_fee_owner"
	MessageUnstakeServicerFeeOwner        = "message_unstake_servicer_fee_owner"
	MessagePauseServicerFeeOwner          = "message_pause_servicer_fee_owner"
	MessageUnpauseServicerFeeOwner        = "message_unpause_servicer_fee_owner"

	MessageChangeParameterFeeOwner = "message_change_parameter_fee_owner"
)
