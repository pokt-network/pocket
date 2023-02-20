package p2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	types "github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/runtime/genesis"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/telemetry"
	"github.com/stretchr/testify/require"
)

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

const (
	serviceUrlFormat  = "node%d.consensus:8080"
	eventsChannelSize = 10000
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

func generateKey(_ *testing.T, seed int) cryptoPocket.PrivateKey {
	seedBytes := make([]byte, ed25519.PrivateKeySize)
	binary.LittleEndian.PutUint32(seedBytes, uint32(seed))
	pk, err := cryptoPocket.NewPrivateKeyFromSeed(seedBytes)
	if err != nil {
		panic(err)
	}
	return pk
}

func generateKeys(_ *testing.T, numValidators int) []cryptoPocket.PrivateKey {
	keys := make([]cryptoPocket.PrivateKey, numValidators)

	for i := range keys {
		seedInt := genesisConfigSeedStart + i
		keys[i] = generateKey(nil, seedInt)
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

func waitForNetworkSimulationCompletion(t *testing.T, wg *sync.WaitGroup) {
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
func createP2PModules(t *testing.T, busMocks []*mockModules.MockBus) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(busMocks))
	for i := range busMocks {
		p2pMod, err := Create(busMocks[i])
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
	copy(valKeys, keys[:numValidators])
	mockGenesisState := createMockGenesisState(valKeys)
	for i := range mockRuntimeMgrs {
		cfg := &configs.Config{
			RootDirectory: "",
			PrivateKey:    valKeys[i].String(),
			P2P: &configs.P2PConfig{
				PrivateKey:     valKeys[i].String(),
				ConsensusPort:  8080,
				UseRainTree:    true,
				ConnectionType: types.ConnectionType_EmptyConnection,
			},
		}

		mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
		mockRuntimeMgr.EXPECT().GetConfig().Return(cfg).AnyTimes()
		mockRuntimeMgr.EXPECT().GetGenesis().Return(mockGenesisState).AnyTimes()
		mockRuntimeMgrs[i] = mockRuntimeMgr
	}
	return mockRuntimeMgrs
}

func createMockBuses(t *testing.T, runtimeMgrs []modules.RuntimeMgr) []*mockModules.MockBus {
	mockBuses := make([]*mockModules.MockBus, len(runtimeMgrs))
	for i := range mockBuses {
		mockBuses[i] = createMockBus(t, runtimeMgrs[i])
	}
	return mockBuses
}

func createMockBus(t *testing.T, runtimeMgr modules.RuntimeMgr) *mockModules.MockBus {
	ctrl := gomock.NewController(t)
	mockBus := mockModules.NewMockBus(ctrl)
	mockBus.EXPECT().GetRuntimeMgr().Return(runtimeMgr).AnyTimes()
	mockBus.EXPECT().RegisterModule(gomock.Any()).DoAndReturn(func(m modules.Module) {
		m.SetBus(mockBus)
	}).AnyTimes()
	mockModulesRegistry := mockModules.NewMockModulesRegistry(ctrl)
	mockModulesRegistry.EXPECT().GetModule(addrbook_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(addrbook_provider.ModuleName)).AnyTimes()
	mockModulesRegistry.EXPECT().GetModule(current_height_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(current_height_provider.ModuleName)).AnyTimes()
	mockBus.EXPECT().GetModulesRegistry().Return(mockModulesRegistry).AnyTimes()
	mockBus.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	return mockBus
}

// createMockGenesisState configures and returns a mocked GenesisState
func createMockGenesisState(valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	genesisState := new(genesis.GenesisState)
	validators := make([]*coreTypes.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		mockActor := &coreTypes.Actor{
			ActorType:       coreTypes.ActorType_ACTOR_TYPE_VAL,
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
func prepareBusMock(busMock *mockModules.MockBus,
	persistenceMock *mockModules.MockPersistenceModule,
	consensusMock *mockModules.MockConsensusModule,
	telemetryMock *mockModules.MockTelemetryModule,
) {
	busMock.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()
}

// Consensus mock - only needed for validatorMap access
func prepareConsensusMock(t *testing.T, busMock *mockModules.MockBus) *mockModules.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	consensusMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	consensusMock.EXPECT().SetBus(busMock).AnyTimes()
	consensusMock.EXPECT().GetModuleName().Return(modules.ConsensusModuleName).AnyTimes()
	busMock.RegisterModule(consensusMock)

	return consensusMock
}

// Persistence mock - only needed for validatorMap access
func preparePersistenceMock(t *testing.T, busMock *mockModules.MockBus, genesisState *genesis.GenesisState) *mockModules.MockPersistenceModule {
	ctrl := gomock.NewController(t)

	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	readContextMock := mockModules.NewMockPersistenceReadContext(ctrl)

	readContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(readContextMock, nil).AnyTimes()
	readContextMock.EXPECT().Close().Return(nil).AnyTimes()

	persistenceMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	persistenceMock.EXPECT().SetBus(busMock).AnyTimes()
	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	busMock.RegisterModule(persistenceMock)

	return persistenceMock
}

// Telemetry mock - Needed to help with proper counts for number of expected network writes
func prepareTelemetryMock(t *testing.T, busMock *mockModules.MockBus, valId string, wg *sync.WaitGroup, expectedNumNetworkWrites int) *mockModules.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := mockModules.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := prepareNoopTimeSeriesAgentMock(t)
	eventMetricsAgentMock := prepareEventMetricsAgentMock(t, valId, wg, expectedNumNetworkWrites)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()

	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()
	telemetryMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	telemetryMock.EXPECT().SetBus(busMock).AnyTimes()
	busMock.RegisterModule(telemetryMock)

	return telemetryMock
}

// Noop mock - no specific business logic to tend to in the timeseries agent mock
func prepareNoopTimeSeriesAgentMock(t *testing.T) *mockModules.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := mockModules.NewMockTimeSeriesAgent(ctrl)

	timeseriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeseriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	return timeseriesAgentMock
}

// Events metric mock - Needed to help with proper counts for number of expected network writes
func prepareEventMetricsAgentMock(t *testing.T, valId string, wg *sync.WaitGroup, expectedNumNetworkWrites int) *mockModules.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mockModules.NewMockEventMetricsAgent(ctrl)

	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e any, l ...any) {
		log.Printf("[valId: %s] Write\n", valId)
		wg.Done()
	}).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).AnyTimes()

	return eventMetricsAgentMock
}

// Network transport mock - used to simulate reading to/from the network via the main events channel
// as well as counting the number of expected reads
func prepareConnMock(t *testing.T, valId string, wg *sync.WaitGroup, expectedNumNetworkReads int) typesP2P.Transport {
	eventsChannel := make(chan []byte, eventsChannelSize)
	ctrl := gomock.NewController(t)
	connMock := mocksP2P.NewMockTransport(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		wg.Done()
		log.Printf("[valId: %s] Read\n", valId)
		data := <-eventsChannel
		return data, nil
	}).Times(expectedNumNetworkReads + 1) // +1 is necessary because there is one extra read of empty data by every channel when it starts

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		eventsChannel <- data
		return nil
	}).Times(expectedNumNetworkReads)

	connMock.EXPECT().Close().Return(nil).Times(1)

	return connMock
}
