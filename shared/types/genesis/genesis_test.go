package genesis

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisFromJson(t *testing.T) {
	genesis := `{
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"validators": [
		  {
			"address": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b",
			"public_key": "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node1.consensus",
			"port": 8080,
			"chains": ["0001", "0021"]
		  }
		]
	  }`

	g := Genesis{}
	err := json.Unmarshal([]byte(genesis), &g)
	require.NoError(t, err)
}
