package pre2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"sync"
	"testing"

	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	genesisConfigSeedStart = 42
)

func TestRainTree(t *testing.T) {
	numValidators := 4
	configs := createConfigs(t, numValidators+1)

	valP2PMods := make([]*p2pModule, numValidators)
	for i := 0; i < numValidators; i++ {
		config := configs[i]
		_ = typesGenesis.GetNodeState(config)
		p2pMod, err := Create(config)
		require.NoError(t, err)
		p2pMod.Start()
		defer p2pMod.Stop()
		valP2PMods[i] = p2pMod.(*p2pModule)
	}

	// Get the config for the P2P module used by the test
	testConfig := configs[len(configs)-1]
	_ = typesGenesis.GetNodeState(testConfig)

	// Start the test's P2P module
	p2pMod, err := Create(testConfig)
	require.NoError(t, err)
	p2pMod.Start()
	defer p2pMod.Stop()

	addrBook := p2pMod.(*p2pModule).network.GetAddrBook()

	var wg sync.WaitGroup
	wg.Add(len(addrBook))

	p := &anypb.Any{}
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)
	// for _, mod := range valP2PMods {}
	// for _, peer := range addrBook {
	// 	go func(peer *typesPre2P.NetworkPeer) {
	// 		defer wg.Done()
	// 		peer.Dialer.Read()
	// 		fmt.Println("Done reading")
	// 	}(peer)
	// }

	// wg.Wait()
}

func createConfigs(t *testing.T, numValidators int) (configs []*config.Config) {
	configs = make([]*config.Config, numValidators)
	for i := range configs {
		seedInt := genesisConfigSeedStart + i
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, uint32(seedInt))
		pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
		require.NoError(t, err)

		configs[i] = &config.Config{
			RootDir: "",
			Genesis: genesisJson(numValidators, seedInt),

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
	return
}

func genesisJson(numValidators, seed int) string {
	return fmt.Sprintf(`{
		"genesis_state_configs": {
			"num_validators": %d,
			"num_applications": 0,
			"num_fisherman": 0,
			"num_servicers": 0,
			"validator_url_format": "val_%%d",
			"keys_seed_start": %d
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash"
	}`, numValidators, seed)
}
