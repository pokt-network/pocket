package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO(team): Define the configs we need and add more tests here.

func TestLoadConfigFromJson(t *testing.T) {
	config := `{
		"root_dir": "/go/src/github.com/pocket-network",
		"private_key": "0000000100000000000000000000000000000000000000000000000000000000b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
		"genesis": "build/config/genesis.json",
		"p2p": {
			"consensus_port": 8080,
			"debug_port": 9080
		},
		"consensus": {
			"pacemaker": {
				"timeout_msec": 5000,
				"manual": true,
				"debug_time_between_steps_msec": 1000
			}
		},
		"persistence": {},
		"utility": {}
	  }`

	c := Config{}
	err := json.Unmarshal([]byte(config), &c)
	require.NoError(t, err)
}
