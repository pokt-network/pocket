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
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

const (
	genesisConfigSeedStart = 42
	maxNumKeys             = 42 // The number of keys generated for all the unit tests. Optimization to avoid regenerating every time.
	serviceUrlFormat       = "val_%d"
	testChannelSize        = 10000
)

// TEST(olshansky): Add configurations tests for dead and partially visible nodes
type TestRainTreeCommConfig map[string]struct {
	numReads  uint16
	numWrites uint16
}

var keys []cryptoPocket.PrivateKey

func init() {
	keys = generateKeys(maxNumKeys)
}

func generateKeys(numValidators int) []cryptoPocket.PrivateKey {
	keys := make([]cryptoPocket.PrivateKey, numValidators)

	for i := range keys {
		seedInt := genesisConfigSeedStart + i
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, uint32(seedInt))
		pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
		if err != nil {
			panic(err)
		}
		keys[i] = pk
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Address().String() < keys[j].Address().String()
	})
	return keys
}

// A mock of the application specific to know if a message was sent to be handeled by the application
func prepareBusMock(t *testing.T, wg *sync.WaitGroup) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)
	busMock.EXPECT().PublishEventToBus(gomock.Any()).Do(func(e *types.PocketEvent) {
		wg.Done()
		fmt.Println("App specific bus mock publishing event to bus")
	}).MaxTimes(1) // Using `MaxTimes` rather than `Times` because originator node implicitly handles the message
	return busMock
}

// The reason with use `MaxTimes` instead of `Times` here is because we could have gotten full coverage
// while a message was still being sent that would have later been dropped due to de-duplication. There
// is a race condition here, but it is okay because our goal is to achieve max coverage with an upper limit
// on the number of expected messages propagated.
func prepareConnMock(t *testing.T, valId string, expectedNumReads, expectedNumWrites uint16) typesPre2P.TransportLayerConn {
	testChannel := make(chan []byte, testChannelSize)
	ctrl := gomock.NewController(t)
	connMock := mocksPre2P.NewMockTransportLayerConn(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		data := <-testChannel
		return data, nil
	}).MaxTimes(int(expectedNumReads + 1)) // INVESTIGATE(olshansky): The +1 is necessary because there is one extra read of empty data by every channel...

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		testChannel <- data
		return nil
	}).MaxTimes(int(expectedNumWrites))

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
		// HACK(olshansky): I hate that we have to do this, but it is outside the scope of this change...
		typesGenesis.ResetNodeState(t)
	}
	return
}

func createConfigs(t *testing.T, numValidators int) (configs []*config.Config) {
	configs = make([]*config.Config, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	validatorConfigs := genesisValidatorConfig(valKeys)

	for i := range configs {
		configs[i] = &config.Config{
			RootDir: "",
			Genesis: genesisJson(numValidators, validatorConfigs),

			PrivateKey: valKeys[i].(cryptoPocket.Ed25519PrivateKey),

			Pre2P: &config.Pre2PConfig{
				ConsensusPort:  8080,
				UseRainTree:    true,
				ConnectionType: config.EmptyConnection,
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

// REFACTOR(olshansky): The fact that we are passing in a genesis string rather than a properly
// configured struct is a big of legacy. Need to fix this sooner rather than later.
func genesisJson(numValidators int, validatorConfigs string) string {
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
	}`, numValidators, genesisConfigSeedStart, validatorConfigs)
}

func genesisValidatorConfig(valKeys []cryptoPocket.PrivateKey) string {
	s := strings.Builder{}
	for i, valKey := range valKeys {
		if i != 0 {
			s.WriteString(",")
		}
		addr := valKey.Address().String()
		s.WriteString(fmt.Sprintf(`{
			"status": 2,
			"service_url": "%s",
			"staked_tokens": "1000000000000000",
			"address": "%s",
			"output": "%s",
			"public_key": "%s"
		}`, validatorId(i+1), addr, addr, valKey.PublicKey().String()))
	}
	return s.String()
}
