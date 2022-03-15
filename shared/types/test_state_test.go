package types

import (
	"log"
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestLoadStateFromGenesis(t *testing.T) {
	pk, err := pcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	cfg := &config.Config{
		Genesis:    genesisJson(),
		PrivateKey: pk.(pcrypto.Ed25519PrivateKey),
	}
	state := GetTestState(cfg)
	require.Equal(t, 4, len(state.ValidatorMap))
}

func TestLoadStateFrompersistence(t *testing.T) {
	log.Println("[TODO] Load state from the persistence module")
}

func genesisJson() string {
	return `
	{
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"consensus_params": {
			"max_mempool_bytes": 50000000,
			"max_block_bytes": 4000000,
			"max_transaction_bytes": 100000
		},
		"validators": [
			{
				"address": "6f1e5b61ed9a821457aa6b4d7c2a2b37715ffb16",
				"public_key": "9be3287795907809407e14439ff198d5bfc7dce6f9bc743cb369146f610b4801",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "node4",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			},
			{
				"address": "db0743e2dcba9ebf2419bde0881beea966689a26",
				"public_key": "dadbd184a2d526f1ebdd5c06fdad9359b228759b4d7f79d66689fa254aad8546",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "node3",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			},
			{
				"address": "e3c1b362c0df36f6b370b8b1479b67dad96392b2",
				"public_key": "6b79c57e6a095239282c04818e96112f3f03a4001ba97a564c23852a3f1ea5fc",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "node2",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			},
			{
				"address": "fa4d86c3b551aa6cd7c3759d040c037ef2c6379f",
				"public_key": "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "1",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			}
		]
	  }`
}
