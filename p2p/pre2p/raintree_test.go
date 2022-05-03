package pre2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	privateKeySeed = 42
)

func TestRainTree(t *testing.T) {
	cfg := createConfig(t, 4)

	_ = typesGenesis.GetNodeState(cfg)

	p2pMod, err := Create(cfg)
	require.NoError(t, err)

	p := &anypb.Any{}
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)

	// return p
}

func createConfig(t *testing.T, numValidators int) *config.Config {
	seed := make([]byte, ed25519.PrivateKeySize)
	binary.LittleEndian.PutUint32(seed, privateKeySeed)
	pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
	require.NoError(t, err)

	return &config.Config{
		RootDir: "",
		Genesis: genesisJson(numValidators),

		PrivateKey: pk.(cryptoPocket.Ed25519PrivateKey),

		Pre2P: &config.Pre2PConfig{
			ConsensusPort:  8080,
			UseRainTree:    true,
			ConnectionType: "pipe",
		},
		P2P:            &config.P2PConfig{},
		Consensus:      &config.ConsensusConfig{},
		PrePersistence: &config.PrePersistenceConfig{},
		Persistence:    &config.PersistenceConfig{},
		Utility:        &config.UtilityConfig{},
	}
}

func genesisJson(numValidators int) string {
	return fmt.Sprintf(`{
		"genesis_state_configs": {
			"num_validators": %d,
			"num_applications": 0,
			"num_fisherman": 0,
			"num_servicers": 0,
			"validator_url_format": "localhost:8080",
			"keys_seed_start": 42
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash"
	}`, numValidators)
}
