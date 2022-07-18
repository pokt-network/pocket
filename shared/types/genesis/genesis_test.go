package genesis

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisStateFromJson(t *testing.T) {
	// "genesis_time": "2022-01-19T00:00:00.000000Z",
	genesis := `{
		"validators": [
		  {
			"address": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b",
			"public_key": "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
			"paused": false,
			"status": 0,
			"service_url": "validator.com",
			"staked_tokens": "42",
			"missed_blocks": 0,
			"paused_height": 0,
			"unstaking_height": 0,
			"output": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b"
		  }
		],
		"accounts": [],
		"pools": [],
		"fisherman": [],
		"service_nodes": [],
		"apps": []
	  }`

	g := GenesisState{}
	err := json.Unmarshal([]byte(genesis), &g)
	require.NoError(t, err)

}
