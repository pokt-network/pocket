package pre2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	mocksPre2P "github.com/pokt-network/pocket/p2p/pre2p/types/mocks"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	genesisConfigSeedStart = 42
	serviceUrlFormat       = "val_%d"
)

// REFACTOR(olshansky): look into refactoring this to use dependency injection with `uber-go/dig` or `uber-go/fx`
func TestRainTree(t *testing.T) {
	// Network configurations
	numValidators := 4
	configs := createConfigs(t, numValidators)

	// Test configurations
	//
	// 	                     val_1
	// 	┌─────────┬────────────┴───────────────┐
	// val_1     val_3                        val_2
	// 			 └───────┐             ┌───────┴───────┐
	// 				   val_4         val_3           val_2
	originatorNode := validatorId(1)
	var expectedCalls = map[string]struct {
		numReads  uint16
		numWrites uint16
	}{
		validatorId(1): {4, 3}, // originator
		validatorId(2): {4, 3}, //
		validatorId(3): {5, 4}, //
		validatorId(4): {3, 2}, //
	}
	var messageHandeledWaitGroup sync.WaitGroup
	messageHandeledWaitGroup.Add(numValidators)

	// Network initialization
	connMocks := make(map[string]typesPre2P.TransportLayerConn)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range expectedCalls {
		connMocks[valId] = prepareConnMock(t, valId, expectedCall.numReads, expectedCall.numWrites)
		busMocks[valId] = prepareBusMock(t, &messageHandeledWaitGroup)
	}

	// Module injection
	p2pModules := prepareP2PModules(t, configs)
	for validatorId, mod := range p2pModules {
		mod.listener = connMocks[validatorId]
		mod.SetBus(busMocks[validatorId])
		for _, peer := range mod.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		mod.Start()
		defer mod.Stop()
	}

	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[originatorNode]
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)

	// Wait for completion
	messageHandeledWaitGroup.Wait()
}

func prepareBusMock(t *testing.T, wg *sync.WaitGroup) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)
	busMock.EXPECT().PublishEventToBus(gomock.Any()).Do(func(e *types.PocketEvent) {
		wg.Done()
		fmt.Println("App specific bus mock publishing event to bus")
	}).Times(1)
	return busMock
}

// The reason with use `MaxTimes` instead of `Times` is because we could have gotten full coverage
// while a message was still being sent that would have later been dropped due to de-duplication. There
// is a race condition here, but it is okay because our goal is max coverage with an upper limit
// on the number of messages propagated.
func prepareConnMock(t *testing.T, valId string, expectedNumReads, expectedNumWrites uint16) typesPre2P.TransportLayerConn {
	testChannel := make(chan []byte, 10000)
	ctrl := gomock.NewController(t)
	connMock := mocksPre2P.NewMockTransportLayerConn(ctrl)
	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		// fmt.Println(valId, "writing")
		// time.Sleep(10 * time.Second)
		testChannel <- data
		return nil
	}).MaxTimes(int(expectedNumWrites))
	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		// fmt.Println(valId, "reading")
		// time.Sleep(10 * time.Second)
		data := <-testChannel
		return data, nil
	}).MaxTimes(int(expectedNumReads))
	connMock.EXPECT().Close().Return(nil).Times(1)
	return connMock
}

func prepareP2PModules(t *testing.T, configs []*config.Config) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(configs))
	for i, config := range configs {
		_ = typesGenesis.GetNodeState(config)
		p2pMod, err := Create(config)
		require.NoError(t, err)
		p2pModules[validatorId(i+1)] = p2pMod.(*p2pModule)
	}
	return
}

func createConfigs(t *testing.T, numValidators int) (configs []*config.Config) {
	configs = make([]*config.Config, numValidators)
	keys := generateKeys(t, numValidators)
	validatorConfigs := genesisValidatorConfig(keys)
	for i := range configs {
		configs[i] = &config.Config{
			RootDir: "",
			Genesis: genesisJson(numValidators, validatorConfigs),

			PrivateKey: keys[i],

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

func validatorId(i int) string {
	return fmt.Sprintf(serviceUrlFormat, i)
}

func generateKeys(t *testing.T, numValidators int) []cryptoPocket.Ed25519PrivateKey {
	keys := make([]cryptoPocket.Ed25519PrivateKey, numValidators)
	for i := range keys {
		seedInt := genesisConfigSeedStart + i
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, uint32(seedInt))
		pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
		require.NoError(t, err)
		keys[i] = pk.(cryptoPocket.Ed25519PrivateKey)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Address().String() < keys[j].Address().String()
	})
	return keys
}

func genesisJson(numVlidators int, validatorConfigs string) string {
	return fmt.Sprintf(`{
		"genesis_state_configs": {
			"num_validators": %d,
			"num_applications": 0,
			"num_fisherman": 0,
			"num_servicers": 0,
			"keys_seed_start": %d
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"validators": [%s]
	}`, numVlidators, genesisConfigSeedStart, validatorConfigs)
}

func genesisValidatorConfig(keys []cryptoPocket.Ed25519PrivateKey) string {
	s := strings.Builder{}
	for i, key := range keys {
		if i != 0 {
			s.WriteString(",")
		}
		addr := key.Address().String()
		s.WriteString(fmt.Sprintf(`{
			"status": 2,
			"service_url": "%s",
			"staked_tokens": "1000000000000000",
			"address": "%s",
			"output": "%s",
			"public_key": "%s"
		}`, validatorId(i+1), addr, addr, key.PublicKey().String()))
	}
	return s.String()
}
