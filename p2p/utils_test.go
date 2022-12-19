package p2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/telemetry"
	"github.com/stretchr/testify/require"
)

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

const (
	serviceUrlFormat = "node%d.consensus:8080"
	testChannelSize  = 10000
	// Since we simulate up to a 27 node network, we will pre-generate a n >= 27 number of keys to avoid generation
	// every time. The genesis config seed start is set for deterministic key generation and 42 was chosen arbitrarily.
	genesisConfigSeedStart = 42
	// Arbitrary value of the number of private keys we should generate during tests so it is only done once
	maxNumKeys = 42
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

	// IMPROVE: A future improvement of these tests could be to specify specifically which
	//          node IDs the specific read or write is coming from or going to.
}

// CLEANUP: This could (should?) be a codebase-wide shared test helper
func validatorId(i int) string {
	return fmt.Sprintf(serviceUrlFormat, i)
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
	case <-time.After(2 * time.Second): // 2 seconds was chosen arbitrarily. In a mocked environment, all messages should finish sending in less than one minute.
		t.Fatal("Timeout waiting for message to be handled")
	}
}

// ~~~~~~ RainTree Unit Test Mocks ~~~~~~

// createP2PModules returns a map of configured p2pModules keyed by an incremental naming convention (eg: `val_1`, `val_2`, etc.)
func createP2PModules(t *testing.T, runtimeConfigs []modules.RuntimeMgr) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(runtimeConfigs))
	for i, runtimeConfig := range runtimeConfigs {
		p2pMod, err := Create(runtimeConfig)
		require.NoError(t, err)
		p2pModules[validatorId(i+1)] = p2pMod.(*p2pModule)
	}
	return
}

// createMockRuntimeMgrs creates `numValidators` instances of mocked `RuntimeMgr` that are essentially
// representing the runtime environments of the validators that we will use in our tests
func createMockRuntimeMgrs(t *testing.T, numValidators int) []modules.RuntimeMgr {
	ctrl := gomock.NewController(t)
	mockRuntimeMgrs := make([]modules.RuntimeMgr, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	mockGenesisState := createMockGenesisState(t, valKeys)
	for i := range mockRuntimeMgrs {
		cfg := &configs.Config{
			RootDirectory: "",
			PrivateKey:    valKeys[i].String(),
			P2P: &configs.P2PConfig{
				PrivateKey:            valKeys[i].String(),
				ConsensusPort:         8080,
				UseRainTree:           true,
				IsEmptyConnectionType: true,
			},
		}

		mockRuntimeMgr := modulesMock.NewMockRuntimeMgr(ctrl)
		mockRuntimeMgr.EXPECT().GetConfig().Return(cfg).AnyTimes()
		mockRuntimeMgr.EXPECT().GetGenesis().Return(mockGenesisState).AnyTimes()
		mockRuntimeMgrs[i] = mockRuntimeMgr
	}
	return mockRuntimeMgrs
}

// createMockGenesisState configures and returns a mocked GenesisState
func createMockGenesisState(t *testing.T, valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	var genesisState = new(genesis.GenesisState)

	validators := make([]*coreTypes.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		mockActor := &coreTypes.Actor{
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
			GenericParam:    validatorId(i + 1),
			StakedAmount:    "1000000000000000",
			PausedHeight:    int64(0),
			UnstakingHeight: int64(0),
			Output:          addr,
		}
		validators[i] = mockActor
	}
	genesisState.Validators = validators

	return genesisState
}

// Bus Mock - needed to return the appropriate modules when accessed
func prepareBusMock(t *testing.T,
	consensusMock *modulesMock.MockConsensusModule,
	persistenceMock *modulesMock.MockPersistenceModule,
	telemetryMock *modulesMock.MockTelemetryModule,
) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)

	busMock.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	return busMock
}

// Consensus mock - only needed for validatorMap access
func prepareConsensusMock(t *testing.T, genesisState *genesis.GenesisState) *modulesMock.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := modulesMock.NewMockConsensusModule(ctrl)

	validators := genesisState.GetValidators()
	m := make(modules.ValidatorMap, len(validators))
	for _, v := range validators {
		m[v.GetAddress()] = *v
	}

	consensusMock.EXPECT().ValidatorMap().Return(m).AnyTimes()
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	return consensusMock
}

// Persistence mock - only needed for validatorMap access
func preparePersistenceMock(t *testing.T, genesisState *genesis.GenesisState) *modulesMock.MockPersistenceModule {
	ctrl := gomock.NewController(t)

	persistenceMock := modulesMock.NewMockPersistenceModule(ctrl)
	readContextMock := modulesMock.NewMockPersistenceReadContext(ctrl)

	readContextMock.EXPECT().GetAllStakedActors(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(readContextMock, nil).AnyTimes()

	return persistenceMock
}

// Telemetry mock - Needed to help with proper counts for number of expected network writes
func prepareTelemetryMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkWrites int) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := prepareNoopTimeSeriesAgentMock(t)
	eventMetricsAgentMock := prepareEventMetricsAgentMock(t, wg, expectedNumNetworkWrites)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()

	return telemetryMock
}

// Noop mock - no specific business logic to tend to in the timeseries agent mock
func prepareNoopTimeSeriesAgentMock(t *testing.T) *modulesMock.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := modulesMock.NewMockTimeSeriesAgent(ctrl)

	timeseriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeseriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	return timeseriesAgentMock
}

// Events metric mock - Needed to help with proper counts for number of expected network writes
func prepareEventMetricsAgentMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkWrites int) *modulesMock.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := modulesMock.NewMockEventMetricsAgent(ctrl)

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e interface{}, l ...interface{}) {
		wg.Done()
	}).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).AnyTimes()

	return eventMetricsAgentMock
}

// Network transport mock - used to simulate reading to/from the network via the main events channel
// as well as counting the number of expected reads
func prepareConnMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkReads int) typesP2P.Transport {
	testChannel := make(chan []byte, testChannelSize)
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
