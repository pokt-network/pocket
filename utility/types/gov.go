package types

// IMPROVE: Rename `UnstakingBlocks` to `UnbondingPeriod` or `UnstakingBlocksUnbondingPeriod`
// IMPROVE: Create a mapping from ActorType to the gov params that are relevant to that actor type.

const (
	// Session gov params
	BlocksPerSessionParamName = "blocks_per_session"

	// Application actor gov params
	AppMinimumStakeParamName       = "app_minimum_stake"
	AppMaxChainsParamName          = "app_max_chains"
	AppUnstakingBlocksParamName    = "app_unstaking_blocks"
	AppMinimumPauseBlocksParamName = "app_minimum_pause_blocks"
	AppMaxPauseBlocksParamName     = "app_max_pause_blocks"
	// The Application's usage tokens during each session is determined by its stake. The session
	// is rate limited using the "Token Bucket" algorithm, where the number of tokens in the beginning
	// of each session is determined by this parameter.
	//nolint:gosec // G101 - Not a hardcoded credential
	AppSessionTokensMultiplierParamName = "app_session_tokens_multiplier"

	// Servicer actor gov params
	ServicerMinimumStakeParamName       = "servicer_minimum_stake"
	ServicerMaxChainsParamName          = "servicer_max_chains"
	ServicerUnstakingBlocksParamName    = "servicer_unstaking_blocks"
	ServicerMinimumPauseBlocksParamName = "servicer_minimum_pause_blocks"
	ServicerMaxPauseBlocksParamName     = "servicer_max_pause_blocks"
	ServicersPerSessionParamName        = "servicers_per_session"

	// Watcher actor gov params
	WatcherMinimumStakeParamName       = "watcher_minimum_stake"
	WatcherMaxChainsParamName          = "watcher_max_chains"
	WatcherUnstakingBlocksParamName    = "watcher_unstaking_blocks"
	WatcherMinimumPauseBlocksParamName = "watcher_minimum_pause_blocks"
	WatcherMaxPauseBlocksParamName     = "watcher_max_pause_blocks"
	WatcherPerSessionParamName         = "watcher_per_session"

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
	MessageStakeWatcherFee         = "message_stake_watcher_fee"
	MessageEditStakeWatcherFee     = "message_edit_stake_watcher_fee"
	MessageUnstakeWatcherFee       = "message_unstake_watcher_fee"
	MessagePauseWatcherFee         = "message_pause_watcher_fee"
	MessageUnpauseWatcherFee       = "message_unpause_watcher_fee"
	MessageWatcherPauseServicerFee = "message_watcher_pause_servicer_fee"
	MessageTestScoreFee            = "message_test_score_fee"
	MessageProveTestScoreFee       = "message_prove_test_score_fee"

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

	AppMinimumStakeOwner = "app_minimum_stake_owner"
	AppMaxChainsOwner    = "app_max_chains_owner"
	//nolint:gosec // G101 - Not a hardcoded credential
	AppSessionTokensMultiplierOwner = "app_session_tokens_multiplier_owner"
	AppUnstakingBlocksOwner         = "app_unstaking_blocks_owner"
	AppMinimumPauseBlocksOwner      = "app_minimum_pause_blocks_owner"
	AppMaxPausedBlocksOwner         = "app_max_paused_blocks_owner"

	ServicerMinimumStakeOwner       = "servicer_minimum_stake_owner"
	ServicerMaxChainsOwner          = "servicer_max_chains_owner"
	ServicerUnstakingBlocksOwner    = "servicer_unstaking_blocks_owner"
	ServicerMinimumPauseBlocksOwner = "servicer_minimum_pause_blocks_owner"
	ServicerMaxPausedBlocksOwner    = "servicer_max_paused_blocks_owner"
	ServicersPerSessionOwner        = "servicers_per_session_owner"

	WatcherMinimumStakeOwner       = "watcher_minimum_stake_owner"
	WatcherMaxChainsOwner          = "watcher_max_chains_owner"
	WatcherUnstakingBlocksOwner    = "watcher_unstaking_blocks_owner"
	WatcherMinimumPauseBlocksOwner = "watcher_minimum_pause_blocks_owner"
	WatcherMaxPausedBlocksOwner    = "watcher_max_paused_blocks_owner"
	WatcherPerSessionOwner         = "watcher_per_session_owner"

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

	MessageStakeWatcherFeeOwner         = "message_stake_watcher_fee_owner"
	MessageEditStakeWatcherFeeOwner     = "message_edit_stake_watcher_fee_owner"
	MessageUnstakeWatcherFeeOwner       = "message_unstake_watcher_fee_owner"
	MessagePauseWatcherFeeOwner         = "message_pause_watcher_fee_owner"
	MessageUnpauseWatcherFeeOwner       = "message_unpause_watcher_fee_owner"
	MessageWatcherPauseServicerFeeOwner = "message_watcher_pause_servicer_fee_owner"
	MessageTestScoreFeeOwner            = "message_test_score_fee_owner"
	MessageProveTestScoreFeeOwner       = "message_prove_test_score_fee_owner"
	MessageStakeAppFeeOwner             = "message_stake_app_fee_owner"
	MessageEditStakeAppFeeOwner         = "message_edit_stake_app_fee_owner"
	MessageUnstakeAppFeeOwner           = "message_unstake_app_fee_owner"
	MessagePauseAppFeeOwner             = "message_pause_app_fee_owner"
	MessageUnpauseAppFeeOwner           = "message_unpause_app_fee_owner"
	MessageStakeValidatorFeeOwner       = "message_stake_validator_fee_owner"
	MessageEditStakeValidatorFeeOwner   = "message_edit_stake_validator_fee_owner"
	MessageUnstakeValidatorFeeOwner     = "message_unstake_validator_fee_owner"
	MessagePauseValidatorFeeOwner       = "message_pause_validator_fee_owner"
	MessageUnpauseValidatorFeeOwner     = "message_unpause_validator_fee_owner"
	MessageStakeServicerFeeOwner        = "message_stake_servicer_fee_owner"
	MessageEditStakeServicerFeeOwner    = "message_edit_stake_servicer_fee_owner"
	MessageUnstakeServicerFeeOwner      = "message_unstake_servicer_fee_owner"
	MessagePauseServicerFeeOwner        = "message_pause_servicer_fee_owner"
	MessageUnpauseServicerFeeOwner      = "message_unpause_servicer_fee_owner"

	MessageChangeParameterFeeOwner = "message_change_parameter_fee_owner"
)
