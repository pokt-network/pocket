package types

import (
	"log"
	"pocket/shared/config"
	"pocket/shared/crypto"
	"testing"
)

func TestLoadStateFromGenesis(t *testing.T) {
	pk, _ := crypto.GeneratePrivateKey()
	cfg := &config.Config{
		Genesis:    genesisJson(),
		PrivateKey: pk.String(),
	}
	state := GetTestState()
	state.LoadStateFromConfig(cfg)
	// require.Equal(t, 4, len(state.ValidatorMap))
}

func TestLoadStateFrompersistence(t *testing.T) {
	log.Println("[TODO] Load state from the persistence module")
}

func genesisJson() string {
	return `
	{
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
		  },
		  {
			"node_id": 2,
			"address": "0273a7327f5cd145ae29a12a76ffbfd4d89c0b78ca247450c05f556c24bc264f",
			"public_key": "6a0f6a283a8e4e86d2a3d60ef9e37ec33f2ab6071a30e0a477735128e7571eb0",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node2.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  },
		  {
			"node_id": 3,
			"address": "2a4156d371f8a49a88a6285e9f2ffd77947eac6801c0cfeccdb79ab4b8705f16",
			"public_key": "ab5696551fe1711c3c31669ff20e1e0bc12cb99917c3ab2412e7c13013dee7e7",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node3.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  },
		  {
			"node_id": 4,
			"address": "ffeb214baf0cc1b8019e91a5e5ba0aa71d58de2cc140dd6885147b5b26299fb8",
			"public_key": "d1f87d985adee0c3466ac0458745998fc0f39a9884897ce4c7548d1db8e10642",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node4.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  }
		]
	  }`
}
