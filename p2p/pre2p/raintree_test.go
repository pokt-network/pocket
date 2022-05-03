package pre2p

import (
	"fmt"
	"net"
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func TestRainTree(t *testing.T) {
	cfg := createConfig()
	fmt.Println(cfg)

	typesGenesis.GetNodeState(cfg)

	p2pMod, err := Create(cfg)
	require.NoError(t, err)

	fmt.Println(p2pMod)

	net.Pipe()
	// How many messages handeled by each node?
	// How many messages sent?
	// How many messages received?

}

func createConfig() *config.Config {
	return &config.Config{
		RootDir: "",
		Genesis: genesisJson(),

		PrivateKey: cryptoPocket.Ed25519PrivateKey([]byte("asdasd")),

		Pre2P: &config.Pre2PConfig{
			ConsensusPort: 8080,
			UseRainTree:   true,
		},
		P2P:            &config.P2PConfig{},
		Consensus:      &config.ConsensusConfig{},
		PrePersistence: &config.PrePersistenceConfig{},
		Persistence:    &config.PersistenceConfig{},
		Utility:        &config.UtilityConfig{},
	}
}

func genesisJson() string {
	return `{
		"genesis_state_configs": {
			"num_validators": 4,
			"num_applications": 0,
			"num_fisherman": 0,
			"num_servicers": 0,
			"validator_url_format": "localhost:8080",
			"keys_seed_start": 42
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash"
	}`
}
