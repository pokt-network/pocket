package genesis

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// func TestGenesisFromJson(t *testing.T) {
// 	genesis := `{
// 		"genesis_time": "2022-01-19T00:00:00.000000Z",
// 		"app_hash": "genesis_block_or_state_hash",
// 		"validators": [
// 		  {
// 			"address": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b",
// 			"public_key": "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
// 			"jailed": false,
// 			"upokt": 5000000000000,
// 			"host": "node1.consensus",
// 			"port": 8080,
// 			"chains": ["0001", "0021"]
// 		  }
// 		]
// 	  }`

// 	g := Genesis{}
// 	err := json.Unmarshal([]byte(genesis), &g)
// 	require.NoError(t, err)
// }

func TestGenesisFromState(t *testing.T) {
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

// req := &proto.JobCreateRequest{}
// err := protojson.Unmarshal(bytes, req)

// Address         HexData `json:"address,omitempty"`
// PublicKey       HexData `json:"public_key,omitempty"`
// Paused          bool    `json:"paused,omitempty"`
// Status          int32   `json:"status,omitempty"`
// ServiceUrl      string  `json:"service_url,omitempty"`
// StakedTokens    string  `json:"staked_tokens,omitempty"`
// MissedBlocks    uint32  `json:"missed_blocks,omitempty"`
// PausedHeight    uint64  `json:"paused_height,omitempty"`
// UnstakingHeight int64   `json:"unstaking_height,omitempty"`
// Output          HexData `json:"output,omitempty"`

// google.protobuf.Timestamp genesis_time = 1;

// repeated Validator validators = 2;
// repeated Account accounts = 3;
// repeated Pool pools = 4;
// repeated Fisherman fishermen = 5;
// repeated ServiceNode service_nodes = 6;
// repeated App apps = 7;

// Params params = 8;
