package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisFromJson(t *testing.T) {
	genesis := `{
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"accounts": [
		  {
			"address": "04c56dfc51c3ec68d90a08a2efaa4b9d3db32b3b",
			"public_key": "03e6b38162ccdd0cd8ed657be73885e0b7b99ca09969729e3390c218cfcff07d",
			"upokt": 69
		  }
		],
		"consensus_params": {
			"max_mempool_bytes": 50000000,

			"max_block_bytes": 4000000,
			"max_transaction_bytes": 100000,

			"vrf_key_refresh_freq_block": 5,
			"vrf_key_validity_block": 5,

			"pace_maker": {
				"timeout_msec": 5000,
				"retry_timeout_msec": 1000,
				"max_timeout_msec": 60000,
				"min_block_freq_msec": 2000
			}
		},
		"validators": [
		  {
			"node_id": 1,
			"address": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b",
			"public_key": "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node1.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  }
		]
	  }`

	g := ConsensusGenesis{}
	err := json.Unmarshal([]byte(genesis), &g)
	require.NoError(t, err)
}
