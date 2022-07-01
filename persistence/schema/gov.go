package schema

import (
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(team): There's obviously a better solution here. Just do it.
const (
	ParamsTableName   = "param"
	ParamsTableSchema = `(
		blocks_per_session INT NOT NULL,
		app_minimum_stake TEXT NOT NULL,
		app_max_chains SMALLINT  NOT NULL,
		app_baseline_stake_rate	 INT NOT NULL,
		app_staking_adjustment INT NOT NULL,
		app_unstaking_blocks SMALLINT NOT NULL,
		app_minimum_pause_blocks SMALLINT NOT NULL,
		app_max_pause_blocks INT NOT NULL,

		service_node_minimum_stake    TEXT NOT NULL,
		service_node_max_chains       SMALLINT NOT NULL,
		service_node_unstaking_blocks INT NOT NULL,
		service_node_minimum_pause_blocks SMALLINT NOT NULL,
		service_node_max_pause_blocks INT NOT NULL,
		service_nodes_per_session SMALLINT NOT NULL,

		fisherman_minimum_stake TEXT NOT NULL,
		fisherman_max_chains SMALLINT NOT NULL,
		fisherman_unstaking_blocks INT NOT NULL,
		fisherman_minimum_pause_blocks SMALLINT NOT NULL,
		fisherman_max_pause_blocks SMALLINT NOT NULL,

		validator_minimum_stake TEXT NOT NULL,
		validator_unstaking_blocks INT NOT NULL,
		validator_minimum_pause_blocks SMALLINT NOT NULL,
		validator_max_pause_blocks SMALLINT NOT NULL,
		validator_maximum_missed_blocks SMALLINT NOT NULL,

		validator_max_evidence_age_in_blocks SMALLINT NOT NULL,
		proposer_percentage_of_fees SMALLINT NOT NULL,
		missed_blocks_burn_percentage SMALLINT NOT NULL,
		double_sign_burn_percentage SMALLINT NOT NULL,

		message_double_sign_fee TEXT NOT NULL,
		message_send_fee TEXT NOT NULL,
		message_stake_fisherman_fee TEXT NOT NULL,
		message_edit_stake_fisherman_fee TEXT NOT NULL,
		message_unstake_fisherman_fee TEXT NOT NULL,
		message_pause_fisherman_fee TEXT NOT NULL,
		message_unpause_fisherman_fee TEXT NOT NULL,
		message_fisherman_pause_service_node_fee TEXT NOT NULL,
		message_test_score_fee TEXT NOT NULL,
		message_prove_test_score_fee TEXT NOT NULL,

		message_stake_app_fee TEXT NOT NULL,
		message_edit_stake_app_fee TEXT NOT NULL,
		message_unstake_app_fee TEXT NOT NULL,
		message_pause_app_fee TEXT NOT NULL,
		message_unpause_app_fee TEXT NOT NULL,
		message_stake_validator_fee TEXT NOT NULL,
		message_edit_stake_validator_fee TEXT NOT NULL,
		message_unstake_validator_fee TEXT NOT NULL,
		message_pause_validator_fee TEXT NOT NULL,
		message_unpause_validator_fee TEXT NOT NULL,
		message_stake_service_node_fee TEXT NOT NULL,
		message_edit_stake_service_node_fee TEXT NOT NULL,
		message_unstake_service_node_fee TEXT NOT NULL,
		message_pause_service_node_fee TEXT NOT NULL,
		message_unpause_service_node_fee TEXT NOT NULL,
		message_change_parameter_fee TEXT NOT NULL,

		acl_owner TEXT NOT NULL,
		blocks_per_session_owner TEXT NOT NULL,
		app_minimum_stake_owner TEXT NOT NULL,
		app_max_chains_owner TEXT NOT NULL,
		app_baseline_stake_rate_owner TEXT NOT NULL,
		app_staking_adjustment_owner TEXT NOT NULL,
		app_unstaking_blocks_owner TEXT NOT NULL,
		app_minimum_pause_blocks_owner TEXT NOT NULL,
		app_max_paused_blocks_owner TEXT NOT NULL,

		service_node_minimum_stake_owner TEXT NOT NULL,
		service_node_max_chains_owner TEXT NOT NULL,
		service_node_unstaking_blocks_owner TEXT NOT NULL,
		service_node_minimum_pause_blocks_owner TEXT NOT NULL,
		service_node_max_paused_blocks_owner TEXT NOT NULL,
		service_nodes_per_session_owner TEXT NOT NULL,
		fisherman_minimum_stake_owner TEXT NOT NULL,
		fisherman_max_chains_owner TEXT NOT NULL,
		fisherman_unstaking_blocks_owner TEXT NOT NULL,
		fisherman_minimum_pause_blocks_owner TEXT NOT NULL,
		fisherman_max_paused_blocks_owner TEXT NOT NULL,
		validator_minimum_stake_owner TEXT NOT NULL,
		validator_unstaking_blocks_owner TEXT NOT NULL,
		validator_minimum_pause_blocks_owner TEXT NOT NULL,
		validator_max_paused_blocks_owner TEXT NOT NULL,
		validator_maximum_missed_blocks_owner TEXT NOT NULL,
		validator_max_evidence_age_in_blocks_owner TEXT NOT NULL,
		proposer_percentage_of_fees_owner TEXT NOT NULL,
		missed_blocks_burn_percentage_owner TEXT NOT NULL,
		double_sign_burn_percentage_owner TEXT NOT NULL,

		message_double_sign_fee_owner TEXT NOT NULL,
		message_send_fee_owner TEXT NOT NULL,
		message_stake_fisherman_fee_owner TEXT NOT NULL,
		message_edit_stake_fisherman_fee_owner TEXT NOT NULL,
		message_unstake_fisherman_fee_owner TEXT NOT NULL,
		message_pause_fisherman_fee_owner TEXT NOT NULL,
		message_unpause_fisherman_fee_owner TEXT NOT NULL,
		message_fisherman_pause_service_node_fee_owner TEXT NOT NULL,
		message_test_score_fee_owner TEXT NOT NULL,
		message_prove_test_score_fee_owner TEXT NOT NULL,
		message_stake_app_fee_owner TEXT NOT NULL,
		message_edit_stake_app_fee_owner TEXT NOT NULL,
		message_unstake_app_fee_owner TEXT NOT NULL,
		message_pause_app_fee_owner TEXT NOT NULL,
		message_unpause_app_fee_owner TEXT NOT NULL,
		message_stake_validator_fee_owner TEXT NOT NULL,
		message_edit_stake_validator_fee_owner TEXT NOT NULL,
		message_unstake_validator_fee_owner TEXT NOT NULL,
		message_pause_validator_fee_owner TEXT NOT NULL,
		message_unpause_validator_fee_owner TEXT NOT NULL,
		message_stake_service_node_fee_owner TEXT NOT NULL,
		message_edit_stake_service_node_fee_owner TEXT NOT NULL,
		message_unstake_service_node_fee_owner TEXT NOT NULL,
		message_pause_service_node_fee_owner TEXT NOT NULL,
		message_unpause_service_node_fee_owner TEXT NOT NULL,
		message_change_parameter_fee_owner TEXT NOT NULL,
		end_height BIGINT NOT NULL
	)`
)

var (
	SQLColumnNames = []string{
		"blocks_per_session",

		"app_minimum_stake",
		"app_max_chains",
		"app_baseline_stake_rate",
		"app_staking_adjustment",
		"app_unstaking_blocks",
		"app_minimum_pause_blocks",
		"app_max_pause_blocks",

		"service_node_minimum_stake",
		"service_node_max_chains",
		"service_node_unstaking_blocks",
		"service_node_minimum_pause_blocks",
		"service_node_max_pause_blocks",
		"service_nodes_per_session",

		"fisherman_minimum_stake",
		"fisherman_max_chains",
		"fisherman_unstaking_blocks",
		"fisherman_minimum_pause_blocks",
		"fisherman_max_pause_blocks",

		"validator_minimum_stake",
		"validator_unstaking_blocks",
		"validator_minimum_pause_blocks",
		"validator_max_pause_blocks",
		"validator_maximum_missed_blocks",

		"validator_max_evidence_age_in_blocks",
		"proposer_percentage_of_fees",
		"missed_blocks_burn_percentage",
		"double_sign_burn_percentage",

		"message_double_sign_fee",
		"message_send_fee",
		"message_stake_fisherman_fee",
		"message_edit_stake_fisherman_fee",
		"message_unstake_fisherman_fee",
		"message_pause_fisherman_fee",
		"message_unpause_fisherman_fee",
		"message_fisherman_pause_service_node_fee",
		"message_test_score_fee",
		"message_prove_test_score_fee",
		"message_stake_app_fee",
		"message_edit_stake_app_fee",
		"message_unstake_app_fee",
		"message_pause_app_fee",
		"message_unpause_app_fee",
		"message_stake_validator_fee",
		"message_edit_stake_validator_fee",
		"message_unstake_validator_fee",
		"message_pause_validator_fee",
		"message_unpause_validator_fee",
		"message_stake_service_node_fee",
		"message_edit_stake_service_node_fee",
		"message_unstake_service_node_fee",
		"message_pause_service_node_fee",
		"message_unpause_service_node_fee",
		"message_change_parameter_fee",

		"acl_owner",
		"blocks_per_session_owner",
		"app_minimum_stake_owner",
		"app_max_chains_owner",
		"app_baseline_stake_rate_owner",
		"app_staking_adjustment_owner",
		"app_unstaking_blocks_owner",
		"app_minimum_pause_blocks_owner",
		"app_max_paused_blocks_owner",
		"service_node_minimum_stake_owner",
		"service_node_max_chains_owner",
		"service_node_unstaking_blocks_owner",
		"service_node_minimum_pause_blocks_owner",
		"service_node_max_paused_blocks_owner",
		"service_nodes_per_session_owner",
		"fisherman_minimum_stake_owner",
		"fisherman_max_chains_owner",
		"fisherman_unstaking_blocks_owner",
		"fisherman_minimum_pause_blocks_owner",
		"fisherman_max_paused_blocks_owner",
		"validator_minimum_stake_owner",
		"validator_unstaking_blocks_owner",
		"validator_minimum_pause_blocks_owner",
		"validator_max_paused_blocks_owner",
		"validator_maximum_missed_blocks_owner",
		"validator_max_evidence_age_in_blocks_owner",
		"proposer_percentage_of_fees_owner",
		"missed_blocks_burn_percentage_owner",
		"double_sign_burn_percentage_owner",
		"message_double_sign_fee_owner",
		"message_send_fee_owner",
		"message_stake_fisherman_fee_owner",
		"message_edit_stake_fisherman_fee_owner",
		"message_unstake_fisherman_fee_owner",
		"message_pause_fisherman_fee_owner",
		"message_unpause_fisherman_fee_owner",
		"message_fisherman_pause_service_node_fee_owner",
		"message_test_score_fee_owner",
		"message_prove_test_score_fee_owner",
		"message_stake_app_fee_owner",
		"message_edit_stake_app_fee_owner",
		"message_unstake_app_fee_owner",
		"message_pause_app_fee_owner",
		"message_unpause_app_fee_owner",
		"message_stake_validator_fee_owner",
		"message_edit_stake_validator_fee_owner",
		"message_unstake_validator_fee_owner",
		"message_pause_validator_fee_owner",
		"message_unpause_validator_fee_owner",
		"message_stake_service_node_fee_owner",
		"message_edit_stake_service_node_fee_owner",
		"message_unstake_service_node_fee_owner",
		"message_pause_service_node_fee_owner",
		"message_unpause_service_node_fee_owner",
		"message_change_parameter_fee_owner",
		"end_height",
	}
)

func InsertParams(params *genesis.Params) string {
	return fmt.Sprintf(`INSERT INTO %s VALUES(%d, '%s', %d, %d, %d, %d, %d, %d,
						'%s',%d,%d,%d,%d,%d,
						'%s',%d,%d,%d,%d,
						'%s',%d,%d,%d,%d,
						%d,%d,%d,%d,
						'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',
						'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',
						'%s','%s','%s','%s','%s','%s','%s','%s','%s',
						'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',
						'%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',%d)`,
		ParamsTableName,
		params.BlocksPerSession,
		params.AppMinimumStake,
		params.AppMaxChains,
		params.AppBaselineStakeRate,
		params.AppStakingAdjustment,
		params.AppUnstakingBlocks,
		params.AppMinimumPauseBlocks,
		params.AppMaxPauseBlocks,

		params.ServiceNodeMinimumStake,
		params.ServiceNodeMaxChains,
		params.ServiceNodeUnstakingBlocks,
		params.ServiceNodeMinimumPauseBlocks,
		params.ServiceNodeMaxPauseBlocks,
		params.ServiceNodesPerSession,

		params.FishermanMinimumStake,
		params.FishermanMaxChains,
		params.FishermanUnstakingBlocks,
		params.FishermanMinimumPauseBlocks,
		params.FishermanMaxPauseBlocks,

		params.ValidatorMinimumStake,
		params.ValidatorUnstakingBlocks,
		params.ValidatorMinimumPauseBlocks,
		params.ValidatorMaxPauseBlocks,
		params.ValidatorMaximumMissedBlocks,

		params.ValidatorMaxEvidenceAgeInBlocks,
		params.ProposerPercentageOfFees,
		params.MissedBlocksBurnPercentage,
		params.DoubleSignBurnPercentage,

		params.MessageDoubleSignFee,
		params.MessageSendFee,
		params.MessageStakeFishermanFee,
		params.MessageEditStakeFishermanFee,
		params.MessageUnstakeFishermanFee,
		params.MessagePauseFishermanFee,
		params.MessageUnpauseFishermanFee,
		params.MessageFishermanPauseServiceNodeFee,
		params.MessageTestScoreFee,
		params.MessageProveTestScoreFee,

		params.MessageStakeAppFee,
		params.MessageEditStakeAppFee,
		params.MessageUnstakeAppFee,
		params.MessagePauseAppFee,
		params.MessageUnpauseAppFee,
		params.MessageStakeValidatorFee,
		params.MessageEditStakeValidatorFee,
		params.MessageUnstakeValidatorFee,
		params.MessagePauseValidatorFee,
		params.MessageUnpauseValidatorFee,
		params.MessageStakeServiceNodeFee,
		params.MessageEditStakeServiceNodeFee,
		params.MessageUnstakeServiceNodeFee,
		params.MessagePauseServiceNodeFee,
		params.MessageUnpauseServiceNodeFee,
		params.MessageChangeParameterFee,

		hex.EncodeToString(params.AclOwner),
		hex.EncodeToString(params.BlocksPerSessionOwner),
		hex.EncodeToString(params.AppMinimumStakeOwner),
		hex.EncodeToString(params.AppMaxChainsOwner),
		hex.EncodeToString(params.AppBaselineStakeRateOwner),
		hex.EncodeToString(params.AppStakingAdjustmentOwner),
		hex.EncodeToString(params.AppUnstakingBlocksOwner),
		hex.EncodeToString(params.AppMinimumPauseBlocksOwner),
		hex.EncodeToString(params.AppMaxPausedBlocksOwner),

		hex.EncodeToString(params.ServiceNodeMinimumStakeOwner),
		hex.EncodeToString(params.ServiceNodeMaxChainsOwner),
		hex.EncodeToString(params.ServiceNodeUnstakingBlocksOwner),
		hex.EncodeToString(params.ServiceNodeMinimumPauseBlocksOwner),
		hex.EncodeToString(params.ServiceNodeMaxPausedBlocksOwner),
		hex.EncodeToString(params.ServiceNodesPerSessionOwner),
		hex.EncodeToString(params.FishermanMinimumStakeOwner),
		hex.EncodeToString(params.FishermanMaxChainsOwner),
		hex.EncodeToString(params.FishermanUnstakingBlocksOwner),
		hex.EncodeToString(params.FishermanMinimumPauseBlocksOwner),
		hex.EncodeToString(params.FishermanMaxPausedBlocksOwner),
		hex.EncodeToString(params.ValidatorMinimumStakeOwner),
		hex.EncodeToString(params.ValidatorUnstakingBlocksOwner),
		hex.EncodeToString(params.ValidatorMinimumPauseBlocksOwner),
		hex.EncodeToString(params.ValidatorMaxPausedBlocksOwner),
		hex.EncodeToString(params.ValidatorMaximumMissedBlocksOwner),
		hex.EncodeToString(params.ValidatorMaxEvidenceAgeInBlocksOwner),
		hex.EncodeToString(params.ProposerPercentageOfFeesOwner),
		hex.EncodeToString(params.MissedBlocksBurnPercentageOwner),
		hex.EncodeToString(params.DoubleSignBurnPercentageOwner),

		hex.EncodeToString(params.MessageDoubleSignFeeOwner),
		hex.EncodeToString(params.MessageSendFeeOwner),
		hex.EncodeToString(params.MessageStakeFishermanFeeOwner),
		hex.EncodeToString(params.MessageEditStakeFishermanFeeOwner),
		hex.EncodeToString(params.MessageUnstakeFishermanFeeOwner),
		hex.EncodeToString(params.MessagePauseFishermanFeeOwner),
		hex.EncodeToString(params.MessageUnpauseFishermanFeeOwner),
		hex.EncodeToString(params.MessageFishermanPauseServiceNodeFeeOwner),
		hex.EncodeToString(params.MessageTestScoreFeeOwner),
		hex.EncodeToString(params.MessageProveTestScoreFeeOwner),
		hex.EncodeToString(params.MessageStakeAppFeeOwner),
		hex.EncodeToString(params.MessageEditStakeAppFeeOwner),
		hex.EncodeToString(params.MessageUnstakeAppFeeOwner),
		hex.EncodeToString(params.MessagePauseAppFeeOwner),
		hex.EncodeToString(params.MessageUnpauseAppFeeOwner),
		hex.EncodeToString(params.MessageStakeValidatorFeeOwner),
		hex.EncodeToString(params.MessageEditStakeValidatorFeeOwner),
		hex.EncodeToString(params.MessageUnstakeValidatorFeeOwner),
		hex.EncodeToString(params.MessagePauseValidatorFeeOwner),
		hex.EncodeToString(params.MessageUnpauseValidatorFeeOwner),
		hex.EncodeToString(params.MessageStakeServiceNodeFeeOwner),
		hex.EncodeToString(params.MessageEditStakeServiceNodeFeeOwner),
		hex.EncodeToString(params.MessageUnstakeServiceNodeFeeOwner),
		hex.EncodeToString(params.MessagePauseServiceNodeFeeOwner),
		hex.EncodeToString(params.MessageUnpauseServiceNodeFeeOwner),
		hex.EncodeToString(params.MessageChangeParameterFeeOwner),
		DefaultEndHeight,
	)
}

func GetParamQuery(paramName string) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE end_height=%d`, paramName, ParamsTableName, DefaultEndHeight)
}

func GetParamNames() (paramNames []string) {
	paramNames = make([]string, len(SQLColumnNames))
	copy(paramNames, SQLColumnNames)
	return
}

func NullifyParamsQuery(height int64) string {
	return fmt.Sprintf(`UPDATE %s SET end_height=%d WHERE end_height=%d`, ParamsTableName, height, DefaultEndHeight)
}

func SetParam(paramName string, paramValue interface{}, height int64) string {
	paramNames := GetParamNames()
	pNamesLen := len(paramNames)
	var index int
	// TODO (Team) optimize linear search
	for i, s := range paramNames {
		if s == paramName {
			index = i
		}
	}
	switch v := paramValue.(type) {
	case int, int32, int64:
		paramNames[index] = fmt.Sprintf("%d", v)
	case []byte:
		paramNames[index] = fmt.Sprintf("'%s'", hex.EncodeToString(v))
	case string:
		paramNames[index] = fmt.Sprintf("'%s'", v)
	default:
		panic("unknown param value")
	}
	subQuery := `SELECT `
	maxIndex := pNamesLen - 1
	for i, pn := range paramNames {
		if i == maxIndex {
			subQuery += "-1"
		} else {
			subQuery += fmt.Sprintf("%s,", pn)
		}
	}
	subQuery += fmt.Sprintf(` FROM %s WHERE end_height=%d`, ParamsTableName, height)
	return fmt.Sprintf(`INSERT INTO %s((%s))`, ParamsTableName, subQuery)
}

func ClearAllGovQuery() string {
	return fmt.Sprintf(`DELETE FROM %s`, ParamsTableName)
}
