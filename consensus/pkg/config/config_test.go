package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromJson(t *testing.T) {
	config := `{
		"root_dir": "/go/src/github.com/pocket-network",
		"private_key": "0000000100000000000000000000000000000000000000000000000000000000b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
		"genesis": "config/genesis.json",
		"p2p": {
		  "consensus_port": 8080,
		  "debug_port": 9080
		},
		"consensus": {
		  "node_id": 1
		},
		"persistence": {},
		"utility": {}
	  }`

	c := Config{}
	err := json.Unmarshal([]byte(config), &c)
	require.NoError(t, err)
}
