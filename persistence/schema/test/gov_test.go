package test

import (
	"testing"

	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

func TestInsertParams(t *testing.T) {
	type args struct {
		params *genesis.Params
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should params from genesis as expected",
			args: args{
				params: &genesis.Params{
					BlocksPerSession:                         1,
					AppMinimumStake:                          "100",
					AppMaxChains:                             2,
					AppBaselineStakeRate:                     3,
					AppStakingAdjustment:                     4,
					AppUnstakingBlocks:                       5,
					AppMinimumPauseBlocks:                    6,
					AppMaxPauseBlocks:                        7,
					ServiceNodeMinimumStake:                  "200",
					ServiceNodeMaxChains:                     8,
					ServiceNodeUnstakingBlocks:               9,
					ServiceNodeMinimumPauseBlocks:            10,
					ServiceNodeMaxPauseBlocks:                11,
					ServiceNodesPerSession:                   12,
					FishermanMinimumStake:                    "300",
					FishermanMaxChains:                       13,
					FishermanUnstakingBlocks:                 14,
					FishermanMinimumPauseBlocks:              15,
					FishermanMaxPauseBlocks:                  16,
					ValidatorMinimumStake:                    "400",
					ValidatorUnstakingBlocks:                 17,
					ValidatorMinimumPauseBlocks:              18,
					ValidatorMaxPauseBlocks:                  19,
					ValidatorMaximumMissedBlocks:             20,
					ValidatorMaxEvidenceAgeInBlocks:          21,
					ProposerPercentageOfFees:                 22,
					MissedBlocksBurnPercentage:               23,
					DoubleSignBurnPercentage:                 24,
					MessageDoubleSignFee:                     "test_MessageDoubleSignFee",
					MessageSendFee:                           "test_MessageSendFee",
					MessageStakeFishermanFee:                 "test_MessageStakeFishermanFee",
					MessageEditStakeFishermanFee:             "test_MessageEditStakeFishermanFee",
					MessageUnstakeFishermanFee:               "test_MessageUnstakeFishermanFee",
					MessagePauseFishermanFee:                 "test_MessagePauseFishermanFee",
					MessageUnpauseFishermanFee:               "test_MessageUnpauseFishermanFee",
					MessageFishermanPauseServiceNodeFee:      "test_MessageFishermanPauseServiceNodeFee",
					MessageTestScoreFee:                      "test_MessageTestScoreFee",
					MessageProveTestScoreFee:                 "test_MessageProveTestScoreFee",
					MessageStakeAppFee:                       "test_MessageStakeAppFee",
					MessageEditStakeAppFee:                   "test_MessageEditStakeAppFee",
					MessageUnstakeAppFee:                     "test_MessageUnstakeAppFee",
					MessagePauseAppFee:                       "test_MessagePauseAppFee",
					MessageUnpauseAppFee:                     "test_MessageUnpauseAppFee",
					MessageStakeValidatorFee:                 "test_MessageStakeValidatorFee",
					MessageEditStakeValidatorFee:             "test_MessageEditStakeValidatorFee",
					MessageUnstakeValidatorFee:               "test_MessageUnstakeValidatorFee",
					MessagePauseValidatorFee:                 "test_MessagePauseValidatorFee",
					MessageUnpauseValidatorFee:               "test_MessageUnpauseValidatorFee",
					MessageStakeServiceNodeFee:               "test_MessageStakeServiceNodeFee",
					MessageEditStakeServiceNodeFee:           "test_MessageEditStakeServiceNodeFee",
					MessageUnstakeServiceNodeFee:             "test_MessageUnstakeServiceNodeFee",
					MessagePauseServiceNodeFee:               "test_MessagePauseServiceNodeFee",
					MessageUnpauseServiceNodeFee:             "test_MessageUnpauseServiceNodeFee",
					MessageChangeParameterFee:                "test_MessageChangeParameterFee",
					AclOwner:                                 []byte("test_AclOwner"),
					BlocksPerSessionOwner:                    []byte("test_BlocksPerSessionOwner"),
					AppMinimumStakeOwner:                     []byte("test_AppMinimumStakeOwner"),
					AppMaxChainsOwner:                        []byte("test_AppMaxChainsOwner"),
					AppBaselineStakeRateOwner:                []byte("test_AppBaselineStakeRateOwner"),
					AppStakingAdjustmentOwner:                []byte("test_AppStakingAdjustmentOwner"),
					AppUnstakingBlocksOwner:                  []byte("test_AppUnstakingBlocksOwner"),
					AppMinimumPauseBlocksOwner:               []byte("test_AppMinimumPauseBlocksOwner"),
					AppMaxPausedBlocksOwner:                  []byte("test_AppMaxPausedBlocksOwner"),
					ServiceNodeMinimumStakeOwner:             []byte("test_ServiceNodeMinimumStakeOwner"),
					ServiceNodeMaxChainsOwner:                []byte("test_ServiceNodeMaxChainsOwner"),
					ServiceNodeUnstakingBlocksOwner:          []byte("test_ServiceNodeUnstakingBlocksOwner"),
					ServiceNodeMinimumPauseBlocksOwner:       []byte("test_ServiceNodeMinimumPauseBlocksOwner"),
					ServiceNodeMaxPausedBlocksOwner:          []byte("test_ServiceNodeMaxPausedBlocksOwner"),
					ServiceNodesPerSessionOwner:              []byte("test_ServiceNodesPerSessionOwner"),
					FishermanMinimumStakeOwner:               []byte("test_FishermanMinimumStakeOwner"),
					FishermanMaxChainsOwner:                  []byte("test_FishermanMaxChainsOwner"),
					FishermanUnstakingBlocksOwner:            []byte("test_FishermanUnstakingBlocksOwner"),
					FishermanMinimumPauseBlocksOwner:         []byte("test_FishermanMinimumPauseBlocksOwner"),
					FishermanMaxPausedBlocksOwner:            []byte("test_FishermanMaxPausedBlocksOwner"),
					ValidatorMinimumStakeOwner:               []byte("test_ValidatorMinimumStakeOwner"),
					ValidatorUnstakingBlocksOwner:            []byte("test_ValidatorUnstakingBlocksOwner"),
					ValidatorMinimumPauseBlocksOwner:         []byte("test_ValidatorMinimumPauseBlocksOwner"),
					ValidatorMaxPausedBlocksOwner:            []byte("test_ValidatorMaxPausedBlocksOwner"),
					ValidatorMaximumMissedBlocksOwner:        []byte("test_ValidatorMaximumMissedBlocksOwner"),
					ValidatorMaxEvidenceAgeInBlocksOwner:     []byte("test_ValidatorMaxEvidenceAgeInBlocksOwner"),
					ProposerPercentageOfFeesOwner:            []byte("test_ProposerPercentageOfFeesOwner"),
					MissedBlocksBurnPercentageOwner:          []byte("test_MissedBlocksBurnPercentageOwner"),
					DoubleSignBurnPercentageOwner:            []byte("test_DoubleSignBurnPercentageOwner"),
					MessageDoubleSignFeeOwner:                []byte("test_MessageDoubleSignFeeOwner"),
					MessageSendFeeOwner:                      []byte("test_MessageSendFeeOwner"),
					MessageStakeFishermanFeeOwner:            []byte("test_MessageStakeFishermanFeeOwner"),
					MessageEditStakeFishermanFeeOwner:        []byte("test_MessageEditStakeFishermanFeeOwner"),
					MessageUnstakeFishermanFeeOwner:          []byte("test_MessageUnstakeFishermanFeeOwner"),
					MessagePauseFishermanFeeOwner:            []byte("test_MessagePauseFishermanFeeOwner"),
					MessageUnpauseFishermanFeeOwner:          []byte("test_MessageUnpauseFishermanFeeOwner"),
					MessageFishermanPauseServiceNodeFeeOwner: []byte("test_MessageFishermanPauseServiceNodeFeeOwner"),
					MessageTestScoreFeeOwner:                 []byte("test_MessageTestScoreFeeOwner"),
					MessageProveTestScoreFeeOwner:            []byte("test_MessageProveTestScoreFeeOwner"),
					MessageStakeAppFeeOwner:                  []byte("test_MessageStakeAppFeeOwner"),
					MessageEditStakeAppFeeOwner:              []byte("test_MessageEditStakeAppFeeOwner"),
					MessageUnstakeAppFeeOwner:                []byte("test_MessageUnstakeAppFeeOwner"),
					MessagePauseAppFeeOwner:                  []byte("test_MessagePauseAppFeeOwner"),
					MessageUnpauseAppFeeOwner:                []byte("test_MessageUnpauseAppFeeOwner"),
					MessageStakeValidatorFeeOwner:            []byte("test_MessageStakeValidatorFeeOwner"),
					MessageEditStakeValidatorFeeOwner:        []byte("test_MessageEditStakeValidatorFeeOwner"),
					MessageUnstakeValidatorFeeOwner:          []byte("test_MessageUnstakeValidatorFeeOwner"),
					MessagePauseValidatorFeeOwner:            []byte("test_MessagePauseValidatorFeeOwner"),
					MessageUnpauseValidatorFeeOwner:          []byte("test_MessageUnpauseValidatorFeeOwner"),
					MessageStakeServiceNodeFeeOwner:          []byte("test_MessageStakeServiceNodeFeeOwner"),
					MessageEditStakeServiceNodeFeeOwner:      []byte("test_MessageEditStakeServiceNodeFeeOwner"),
					MessageUnstakeServiceNodeFeeOwner:        []byte("test_MessageUnstakeServiceNodeFeeOwner"),
					MessagePauseServiceNodeFeeOwner:          []byte("test_MessagePauseServiceNodeFeeOwner"),
					MessageUnpauseServiceNodeFeeOwner:        []byte("test_MessageUnpauseServiceNodeFeeOwner"),
					MessageChangeParameterFeeOwner:           []byte("test_MessageChangeParameterFeeOwner"),
				},
			},
			want: "INSERT INTO params VALUES ('blocks_per_session', -1, true, 'BIGINT', 1),('app_minimum_stake', -1, true, 'STRING', '100'),('app_max_chains', -1, true, 'SMALLINT', 2),('app_baseline_stake_rate', -1, true, 'BIGINT', 3),('app_staking_adjustment', -1, true, 'BIGINT', 4),('app_unstaking_blocks', -1, true, 'BIGINT', 5),('app_minimum_pause_blocks', -1, true, 'SMALLINT', 6),('app_max_pause_blocks', -1, true, 'BIGINT', 7),('service_node_minimum_stake', -1, true, 'STRING', '200'),('service_node_max_chains', -1, true, 'SMALLINT', 8),('service_node_unstaking_blocks', -1, true, 'BIGINT', 9),('service_node_minimum_pause_blocks', -1, true, 'SMALLINT', 10),('service_node_max_pause_blocks', -1, true, 'BIGINT', 11),('service_nodes_per_session', -1, true, 'SMALLINT', 12),('fisherman_minimum_stake', -1, true, 'STRING', '300'),('fisherman_max_chains', -1, true, 'SMALLINT', 13),('fisherman_unstaking_blocks', -1, true, 'BIGINT', 14),('fisherman_minimum_pause_blocks', -1, true, 'SMALLINT', 15),('fisherman_max_pause_blocks', -1, true, 'SMALLINT', 16),('validator_minimum_stake', -1, true, 'STRING', '400'),('validator_unstaking_blocks', -1, true, 'BIGINT', 17),('validator_minimum_pause_blocks', -1, true, 'SMALLINT', 18),('validator_max_pause_blocks', -1, true, 'SMALLINT', 19),('validator_maximum_missed_blocks', -1, true, 'SMALLINT', 20),('validator_max_evidence_age_in_blocks', -1, true, 'SMALLINT', 21),('proposer_percentage_of_fees', -1, true, 'SMALLINT', 22),('missed_blocks_burn_percentage', -1, true, 'SMALLINT', 23),('double_sign_burn_percentage', -1, true, 'SMALLINT', 24),('message_double_sign_fee', -1, true, 'STRING', 'test_MessageDoubleSignFee'),('message_send_fee', -1, true, 'STRING', 'test_MessageSendFee'),('message_stake_fisherman_fee', -1, true, 'STRING', 'test_MessageStakeFishermanFee'),('message_edit_stake_fisherman_fee', -1, true, 'STRING', 'test_MessageEditStakeFishermanFee'),('message_unstake_fisherman_fee', -1, true, 'STRING', 'test_MessageUnstakeFishermanFee'),('message_pause_fisherman_fee', -1, true, 'STRING', 'test_MessagePauseFishermanFee'),('message_unpause_fisherman_fee', -1, true, 'STRING', 'test_MessageUnpauseFishermanFee'),('message_fisherman_pause_service_node_fee', -1, true, 'STRING', 'test_MessageFishermanPauseServiceNodeFee'),('message_test_score_fee', -1, true, 'STRING', 'test_MessageTestScoreFee'),('message_prove_test_score_fee', -1, true, 'STRING', 'test_MessageProveTestScoreFee'),('message_stake_app_fee', -1, true, 'STRING', 'test_MessageStakeAppFee'),('message_edit_stake_app_fee', -1, true, 'STRING', 'test_MessageEditStakeAppFee'),('message_unstake_app_fee', -1, true, 'STRING', 'test_MessageUnstakeAppFee'),('message_pause_app_fee', -1, true, 'STRING', 'test_MessagePauseAppFee'),('message_unpause_app_fee', -1, true, 'STRING', 'test_MessageUnpauseAppFee'),('message_stake_validator_fee', -1, true, 'STRING', 'test_MessageStakeValidatorFee'),('message_edit_stake_validator_fee', -1, true, 'STRING', 'test_MessageEditStakeValidatorFee'),('message_unstake_validator_fee', -1, true, 'STRING', 'test_MessageUnstakeValidatorFee'),('message_pause_validator_fee', -1, true, 'STRING', 'test_MessagePauseValidatorFee'),('message_unpause_validator_fee', -1, true, 'STRING', 'test_MessageUnpauseValidatorFee'),('message_stake_service_node_fee', -1, true, 'STRING', 'test_MessageStakeServiceNodeFee'),('message_edit_stake_service_node_fee', -1, true, 'STRING', 'test_MessageEditStakeServiceNodeFee'),('message_unstake_service_node_fee', -1, true, 'STRING', 'test_MessageUnstakeServiceNodeFee'),('message_pause_service_node_fee', -1, true, 'STRING', 'test_MessagePauseServiceNodeFee'),('message_unpause_service_node_fee', -1, true, 'STRING', 'test_MessageUnpauseServiceNodeFee'),('message_change_parameter_fee', -1, true, 'STRING', 'test_MessageChangeParameterFee'),('acl_owner', -1, true, 'STRING', 'test_AclOwner'),('blocks_per_session_owner', -1, true, 'STRING', 'test_BlocksPerSessionOwner'),('app_minimum_stake_owner', -1, true, 'STRING', 'test_AppMinimumStakeOwner'),('app_max_chains_owner', -1, true, 'STRING', 'test_AppMaxChainsOwner'),('app_baseline_stake_rate_owner', -1, true, 'STRING', 'test_AppBaselineStakeRateOwner'),('app_staking_adjustment_owner', -1, true, 'STRING', 'test_AppStakingAdjustmentOwner'),('app_unstaking_blocks_owner', -1, true, 'STRING', 'test_AppUnstakingBlocksOwner'),('app_minimum_pause_blocks_owner', -1, true, 'STRING', 'test_AppMinimumPauseBlocksOwner'),('app_max_paused_blocks_owner', -1, true, 'STRING', 'test_AppMaxPausedBlocksOwner'),('service_node_minimum_stake_owner', -1, true, 'STRING', 'test_ServiceNodeMinimumStakeOwner'),('service_node_max_chains_owner', -1, true, 'STRING', 'test_ServiceNodeMaxChainsOwner'),('service_node_unstaking_blocks_owner', -1, true, 'STRING', 'test_ServiceNodeUnstakingBlocksOwner'),('service_node_minimum_pause_blocks_owner', -1, true, 'STRING', 'test_ServiceNodeMinimumPauseBlocksOwner'),('service_node_max_paused_blocks_owner', -1, true, 'STRING', 'test_ServiceNodeMaxPausedBlocksOwner'),('service_nodes_per_session_owner', -1, true, 'STRING', 'test_ServiceNodesPerSessionOwner'),('fisherman_minimum_stake_owner', -1, true, 'STRING', 'test_FishermanMinimumStakeOwner'),('fisherman_max_chains_owner', -1, true, 'STRING', 'test_FishermanMaxChainsOwner'),('fisherman_unstaking_blocks_owner', -1, true, 'STRING', 'test_FishermanUnstakingBlocksOwner'),('fisherman_minimum_pause_blocks_owner', -1, true, 'STRING', 'test_FishermanMinimumPauseBlocksOwner'),('fisherman_max_paused_blocks_owner', -1, true, 'STRING', 'test_FishermanMaxPausedBlocksOwner'),('validator_minimum_stake_owner', -1, true, 'STRING', 'test_ValidatorMinimumStakeOwner'),('validator_unstaking_blocks_owner', -1, true, 'STRING', 'test_ValidatorUnstakingBlocksOwner'),('validator_minimum_pause_blocks_owner', -1, true, 'STRING', 'test_ValidatorMinimumPauseBlocksOwner'),('validator_max_paused_blocks_owner', -1, true, 'STRING', 'test_ValidatorMaxPausedBlocksOwner'),('validator_maximum_missed_blocks_owner', -1, true, 'STRING', 'test_ValidatorMaximumMissedBlocksOwner'),('validator_max_evidence_age_in_blocks_owner', -1, true, 'STRING', 'test_ValidatorMaxEvidenceAgeInBlocksOwner'),('proposer_percentage_of_fees_owner', -1, true, 'STRING', 'test_ProposerPercentageOfFeesOwner'),('missed_blocks_burn_percentage_owner', -1, true, 'STRING', 'test_MissedBlocksBurnPercentageOwner'),('double_sign_burn_percentage_owner', -1, true, 'STRING', 'test_DoubleSignBurnPercentageOwner'),('message_double_sign_fee_owner', -1, true, 'STRING', 'test_MessageDoubleSignFeeOwner'),('message_send_fee_owner', -1, true, 'STRING', 'test_MessageSendFeeOwner'),('message_stake_fisherman_fee_owner', -1, true, 'STRING', 'test_MessageStakeFishermanFeeOwner'),('message_edit_stake_fisherman_fee_owner', -1, true, 'STRING', 'test_MessageEditStakeFishermanFeeOwner'),('message_unstake_fisherman_fee_owner', -1, true, 'STRING', 'test_MessageUnstakeFishermanFeeOwner'),('message_pause_fisherman_fee_owner', -1, true, 'STRING', 'test_MessagePauseFishermanFeeOwner'),('message_unpause_fisherman_fee_owner', -1, true, 'STRING', 'test_MessageUnpauseFishermanFeeOwner'),('message_fisherman_pause_service_node_fee_owner', -1, true, 'STRING', 'test_MessageFishermanPauseServiceNodeFeeOwner'),('message_test_score_fee_owner', -1, true, 'STRING', 'test_MessageTestScoreFeeOwner'),('message_prove_test_score_fee_owner', -1, true, 'STRING', 'test_MessageProveTestScoreFeeOwner'),('message_stake_app_fee_owner', -1, true, 'STRING', 'test_MessageStakeAppFeeOwner'),('message_edit_stake_app_fee_owner', -1, true, 'STRING', 'test_MessageEditStakeAppFeeOwner'),('message_unstake_app_fee_owner', -1, true, 'STRING', 'test_MessageUnstakeAppFeeOwner'),('message_pause_app_fee_owner', -1, true, 'STRING', 'test_MessagePauseAppFeeOwner'),('message_unpause_app_fee_owner', -1, true, 'STRING', 'test_MessageUnpauseAppFeeOwner'),('message_stake_validator_fee_owner', -1, true, 'STRING', 'test_MessageStakeValidatorFeeOwner'),('message_edit_stake_validator_fee_owner', -1, true, 'STRING', 'test_MessageEditStakeValidatorFeeOwner'),('message_unstake_validator_fee_owner', -1, true, 'STRING', 'test_MessageUnstakeValidatorFeeOwner'),('message_pause_validator_fee_owner', -1, true, 'STRING', 'test_MessagePauseValidatorFeeOwner'),('message_unpause_validator_fee_owner', -1, true, 'STRING', 'test_MessageUnpauseValidatorFeeOwner'),('message_stake_service_node_fee_owner', -1, true, 'STRING', 'test_MessageStakeServiceNodeFeeOwner'),('message_edit_stake_service_node_fee_owner', -1, true, 'STRING', 'test_MessageEditStakeServiceNodeFeeOwner'),('message_unstake_service_node_fee_owner', -1, true, 'STRING', 'test_MessageUnstakeServiceNodeFeeOwner'),('message_pause_service_node_fee_owner', -1, true, 'STRING', 'test_MessagePauseServiceNodeFeeOwner'),('message_unpause_service_node_fee_owner', -1, true, 'STRING', 'test_MessageUnpauseServiceNodeFeeOwner'),('message_change_parameter_fee_owner', -1, true, 'STRING', 'test_MessageChangeParameterFeeOwner')",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := schema.InsertParams(tt.args.params); got != tt.want {
				t.Errorf("InsertParams() = %v, want %v", got, tt.want)
			}
		})
	}

}

// func TestSetParam(t *testing.T) {
// 	t.Run("int", func(t *testing.T) {
// 		testSetParam[int](t)
// 	})
// 	t.Run("int32", func(t *testing.T) {
// 		testSetParam[int32](t)
// 	})
// 	t.Run("int64", func(t *testing.T) {
// 		testSetParam[int64](t)
// 	})
// 	t.Run("[]byte", func(t *testing.T) {
// 		testSetParam[[]byte](t)
// 	})
// 	t.Run("string", func(t *testing.T) {
// 		testSetParam[string](t)
// 	})

// }

// func testSetParam[T schema.ParamTypes](t *testing.T) {
// 	type args struct {
// 		paramName  string
// 		paramValue T
// 		height     int64
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := schema.SetParam(tt.args.paramName, tt.args.paramValue, tt.args.height); got != tt.want {
// 				t.Errorf("SetParam() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
