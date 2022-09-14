package p2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	p2pTelemetry "github.com/pokt-network/pocket/p2p/telemetry"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	"github.com/stretchr/testify/require"
)

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

// TODO(drewsky): We should leverage the shared `test_artifacts` package and replace the code below.
//                The reason it is necessary right now is because the funcionality of RainTree is dependant
//                on the order of the addresses, which is a function of the public key, so we need to make
//                sure that the validatorId order corresponds to that of the public keys.

const (
	// TODO (olshansk) explain these values
	// Attempt: Since we simulate up to a 27 node network, we will pre-generate a n >= 27 number of keys to avoid generation everytime
	// The genesis config seed start must begin at the max_keys value because...? and 42 is chosen because...?
	// Arbitrary value to use as the seed for deterministic private key generation since RainTree is dependant on lexographic address order
	genesisConfigSeedStart = 42
	// Arbitrary value of the number of private keys we should generate during tests so it is only done once
	maxNumKeys = 42 // Arbitrary number
)

var keys []cryptoPocket.PrivateKey

func init() {
	keys = generateKeys(nil, maxNumKeys)
}

func generateKeys(_ *testing.T, numValidators int) []cryptoPocket.PrivateKey {
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

// A configuration helper used to specify how many messages are expected to be sent or read by the
// P2P module over the network.
type TestNetworkSimulationConfig map[string]struct {
	// The number of asynchronous reads the node's P2P listener made (i.e. # of messages it received over the network)
	numNetworkReads int
	// The number of asynchronous writes the node's P2P listener made (i.e. # of messages it tried to send over the network)
	numNetworkWrites int
	// IMPROVE(drewsky): A future improvement of these tests could be to specify specifically which
	//                   node IDs the specific read or write is coming from or going to.
}

// CLEANUP: This could (should?) be a codebase-wide shared test helper
func validatorId(i int) string {
	return test_artifacts.GetServiceUrl(i)
}

func prepareP2PModulesWithMocks(t *testing.T, networkSimulationConfig TestNetworkSimulationConfig, wg *sync.WaitGroup) map[string]*p2pModule {
	// Network configurations
	numValidators := len(networkSimulationConfig)
	configs, genesisState := createConfigs(t, numValidators)

	// Network & module mocks
	connMocks := make(map[string]typesP2P.Transport)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range networkSimulationConfig {
		wg.Add(expectedCall.numNetworkReads + 1)
		connMocks[valId] = prepareConnMock(t, wg, expectedCall.numNetworkReads)

		wg.Add(expectedCall.numNetworkWrites)
		consensusMock := prepareConsensusMock(t, genesisState)
		telemetryMock := prepareTelemetryMock(t, wg, expectedCall.numNetworkWrites)
		busMocks[valId] = prepareBusMock(t, configs[valId], consensusMock, telemetryMock)
	}

	// Create test P2P Modules
	p2pModules := make(map[string]*p2pModule, len(configs))
	for valId, config := range configs {
		// Create a real P2P module
		p2pMod, err := Create(config, nil)
		require.NoError(t, err)

		// Configure the module's listener's and bus
		p2pModule := p2pMod.(*p2pModule)
		p2pModule.listener = connMocks[valId]
		p2pModule.SetBus(busMocks[valId])

		// Configure how the P2P module communicates "over the network" by setting the appropriate connection mocks
		require.NoError(t, p2pModule.Start()) // Needs to be started before the AddrBook is accessible
		for _, peer := range p2pModule.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		p2pModules[valId] = p2pModule

	}
	return p2pModules
}

func waitForNetworkSimulationCompletion(t *testing.T, p2pModules map[string]*p2pModule, wg *sync.WaitGroup) {
	// Wait for all messages to be transmitted
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Timeout or succeed
	select {
	case <-done:
	// All done!
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message to be handled")
	}

	// Stop all the P2P modules
	for _, p2pModule := range p2pModules {
		p2pModule.Stop()
	}
}

// CLEANUP: Delete this function and use the helpers in `test_artifacts` once we have support for
//          deterministic or injected (for the purpose of ordering) private keys.
func createConfigs(t *testing.T, numValidators int) (configs map[string]*genesis.Config, genesisState *genesis.GenesisState) {
	configs = make(map[string]*genesis.Config, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	genesisState = createGenesisState(t, valKeys)

	for i := 0; i < numValidators; i++ {
		configs[validatorId(i+1)] = &genesis.Config{
			Base: &genesis.BaseConfig{
				RootDirectory: "",
				PrivateKey:    valKeys[i].String(),
			},
			Consensus:   &genesis.ConsensusConfig{},
			Utility:     &genesis.UtilityConfig{},
			Persistence: &genesis.PersistenceConfig{},
			P2P: &genesis.P2PConfig{
				ConsensusPort:  8080,
				UseRainTree:    true,
				ConnectionType: genesis.ConnectionType_EmptyConnection,
			},
			Telemetry: nil,
		}
	}
	return
}

// CLEANUP: Delete this function and use the helpers in `test_artifacts` once we have support for
//          deterministic or injected (for the purpose of ordering) private keys.
func createGenesisState(_ *testing.T, valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	validators := make([]*genesis.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		val := &genesis.Actor{
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
			GenericParam:    validatorId(i + 1),
			StakedAmount:    "1000000000000000",
			PausedHeight:    0,
			UnstakingHeight: 0,
			Output:          addr,
		}
		validators[i] = val
	}
	return &genesis.GenesisState{
		Utility: &genesis.UtilityGenesisState{
			Validators: validators,
		},
	}
}

// ~~~~~~ RainTree Unit Test Mocks ~~~~~~

// A mock of the application specific to know if a message was sent to be handled by the application
func prepareBusMock(t *testing.T, config *genesis.Config, consensusMock *modulesMock.MockConsensusModule,
	telemetryMock *modulesMock.MockTelemetryModule) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)

	busMock.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	busMock.EXPECT().GetConfig().Return(config).Times(1)
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	return busMock
}

// TODO (Olshansk) explain what the consensus mock does (ditto for all mocks below)
// Attempt: the consensus mock returns the genesis validator map anytime the .ValidatorMap() function is called
// the consensus mock also returns '1' when the current height is called

// Consensus mocked - only needed for validatorMap access
func prepareConsensusMock(t *testing.T, genesisState *genesis.GenesisState) *modulesMock.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := modulesMock.NewMockConsensusModule(ctrl)

	validators := genesisState.Utility.Validators
	m := make(modules.ValidatorMap, len(validators))
	for _, v := range validators {
		m[v.Address] = v
	}

	consensusMock.EXPECT().ValidatorMap().Return(m).AnyTimes()
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
	return consensusMock
}

// TODO(team): make the test more rigorous but adding MaxTimes `EmitEvent` expectations. Since we are talking about more than one node
// I have decided to do with `AnyTimes` for the moment.
// TODO (Olshansk) explain why it's necessary to mock telemetry here and document the sub telemetry package
func prepareTelemetryMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkWrites int) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := prepareTimeSeriesAgentMock(t)
	eventMetricsAgentMock := prepareEventMetricsAgentMock(t)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	// timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	// timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(p2pTelemetry.RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e interface{}, l ...interface{}) {
		wg.Done()
	}).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(p2pTelemetry.RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).AnyTimes()

	return telemetryMock
}

// Noop mock - no specific business logic to tend to
func prepareNoopTimeSeriesAgentMock(t *testing.T) *modulesMock.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := modulesMock.NewMockTimeSeriesAgent(ctrl)

	timeseriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeseriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeseriesAgentMock
}

func prepareEventMetricsAgentMock(t *testing.T) *modulesMock.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := modulesMock.NewMockEventMetricsAgent(ctrl)
	return eventMetricsAgentMock
}

func prepareConnMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkReads int) typesP2P.Transport {
	testChannel := make(chan []byte, 1000)
	ctrl := gomock.NewController(t)
	connMock := mocksP2P.NewMockTransport(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		wg.Done()
		data := <-testChannel
		return data, nil
	}).Times(expectedNumNetworkReads + 1) // +1 is necessary because there is one extra read of empty data by every channel when it starts

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		testChannel <- data
		return nil
	}).Times(expectedNumNetworkReads)

	connMock.EXPECT().Close().Return(nil).Times(1)

	return connMock
}
