package types

import (
	"testing"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
)

func TestInsertParams(t *testing.T) {
	type args struct {
		params *genesis.Params
		height int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should insert genesis.DefaultParams() as expected",
			args: args{
				params: test_artifacts.DefaultParams(),
				height: DefaultBigInt,
			},
			want: "INSERT INTO params VALUES ('blocks_per_session', -1, 'BIGINT', 4)," +
				"('app_minimum_stake', -1, 'STRING', '15000000000')," +
				"('app_max_chains', -1, 'SMALLINT', 15)," +
				"('app_baseline_stake_rate', -1, 'BIGINT', 100)," +
				"('app_staking_adjustment', -1, 'BIGINT', 0)," +
				"('app_unstaking_blocks', -1, 'BIGINT', 2016)," +
				"('app_minimum_pause_blocks', -1, 'SMALLINT', 4)," +
				"('app_max_pause_blocks', -1, 'BIGINT', 672)," +
				"('servicer_minimum_stake', -1, 'STRING', '15000000000')," +
				"('servicer_max_chains', -1, 'SMALLINT', 15)," +
				"('servicer_unstaking_blocks', -1, 'BIGINT', 2016)," +
				"('servicer_minimum_pause_blocks', -1, 'SMALLINT', 4)," +
				"('servicer_max_pause_blocks', -1, 'BIGINT', 672)," +
				"('servicers_per_session', -1, 'SMALLINT', 24)," +
				"('fisherman_minimum_stake', -1, 'STRING', '15000000000')," +
				"('fisherman_max_chains', -1, 'SMALLINT', 15)," +
				"('fisherman_unstaking_blocks', -1, 'BIGINT', 2016)," +
				"('fisherman_minimum_pause_blocks', -1, 'SMALLINT', 4)," +
				"('fisherman_max_pause_blocks', -1, 'SMALLINT', 672)," +
				"('validator_minimum_stake', -1, 'STRING', '15000000000')," +
				"('validator_unstaking_blocks', -1, 'BIGINT', 2016)," +
				"('validator_minimum_pause_blocks', -1, 'SMALLINT', 4)," +
				"('validator_max_pause_blocks', -1, 'SMALLINT', 672)," +
				"('validator_maximum_missed_blocks', -1, 'SMALLINT', 5)," +
				"('validator_max_evidence_age_in_blocks', -1, 'SMALLINT', 8)," +
				"('proposer_percentage_of_fees', -1, 'SMALLINT', 10)," +
				"('missed_blocks_burn_percentage', -1, 'SMALLINT', 1)," +
				"('double_sign_burn_percentage', -1, 'SMALLINT', 5)," +
				"('message_double_sign_fee', -1, 'STRING', '10000')," +
				"('message_send_fee', -1, 'STRING', '10000')," +
				"('message_stake_fisherman_fee', -1, 'STRING', '10000')," +
				"('message_edit_stake_fisherman_fee', -1, 'STRING', '10000')," +
				"('message_unstake_fisherman_fee', -1, 'STRING', '10000')," +
				"('message_pause_fisherman_fee', -1, 'STRING', '10000')," +
				"('message_unpause_fisherman_fee', -1, 'STRING', '10000')," +
				"('message_fisherman_pause_servicer_fee', -1, 'STRING', '10000')," +
				"('message_test_score_fee', -1, 'STRING', '10000')," +
				"('message_prove_test_score_fee', -1, 'STRING', '10000')," +
				"('message_stake_app_fee', -1, 'STRING', '10000')," +
				"('message_edit_stake_app_fee', -1, 'STRING', '10000')," +
				"('message_unstake_app_fee', -1, 'STRING', '10000')," +
				"('message_pause_app_fee', -1, 'STRING', '10000')," +
				"('message_unpause_app_fee', -1, 'STRING', '10000')," +
				"('message_stake_validator_fee', -1, 'STRING', '10000')," +
				"('message_edit_stake_validator_fee', -1, 'STRING', '10000')," +
				"('message_unstake_validator_fee', -1, 'STRING', '10000')," +
				"('message_pause_validator_fee', -1, 'STRING', '10000')," +
				"('message_unpause_validator_fee', -1, 'STRING', '10000')," +
				"('message_stake_servicer_fee', -1, 'STRING', '10000')," +
				"('message_edit_stake_servicer_fee', -1, 'STRING', '10000')," +
				"('message_unstake_servicer_fee', -1, 'STRING', '10000')," +
				"('message_pause_servicer_fee', -1, 'STRING', '10000')," +
				"('message_unpause_servicer_fee', -1, 'STRING', '10000')," +
				"('message_change_parameter_fee', -1, 'STRING', '10000')," +
				"('acl_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('blocks_per_session_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_minimum_stake_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_max_chains_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_baseline_stake_rate_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_staking_adjustment_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_unstaking_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_minimum_pause_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('app_max_paused_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('servicer_minimum_stake_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('servicer_max_chains_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('servicer_unstaking_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('servicer_minimum_pause_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('servicer_max_paused_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('servicers_per_session_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('fisherman_minimum_stake_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('fisherman_max_chains_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('fisherman_unstaking_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('fisherman_minimum_pause_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('fisherman_max_paused_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('validator_minimum_stake_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('validator_unstaking_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('validator_minimum_pause_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('validator_max_paused_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('validator_maximum_missed_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('validator_max_evidence_age_in_blocks_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('proposer_percentage_of_fees_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('missed_blocks_burn_percentage_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('double_sign_burn_percentage_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_double_sign_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_send_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_stake_fisherman_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_edit_stake_fisherman_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unstake_fisherman_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_pause_fisherman_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unpause_fisherman_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_fisherman_pause_servicer_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_test_score_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_prove_test_score_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_stake_app_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_edit_stake_app_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unstake_app_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_pause_app_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unpause_app_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_stake_validator_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_edit_stake_validator_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unstake_validator_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_pause_validator_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unpause_validator_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_stake_servicer_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_edit_stake_servicer_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unstake_servicer_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_pause_servicer_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_unpause_servicer_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45')," +
				"('message_change_parameter_fee_owner', -1, 'STRING', 'da034209758b78eaea06dd99c07909ab54c99b45') " +
				"ON CONFLICT ON CONSTRAINT params_pkey DO UPDATE SET value=EXCLUDED.value, type=EXCLUDED.type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InsertParams(tt.args.params, tt.args.height); got != tt.want {
				t.Errorf("InsertParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
