package p2p

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mock_types "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/telemetry"
)

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

const (
	// TECHDEBT: Look into ways to remove `serviceURLFormat` from the test suite
	serviceURLFormat = "node%d.consensus:42069"
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
		keys[i] = cryptoPocket.GetPrivKeySeed(seedInt)
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
// TECHDEBT: rename `validatorId()` to `serviceURL()`
func validatorId(i int) string {
	return fmt.Sprintf(serviceURLFormat, i)
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
func createP2PModules(t *testing.T, busMocks []*mockModules.MockBus, netMock mocknet.Mocknet) (p2pModules map[string]*p2pModule) {
	peerIDs := setupMockNetPeers(t, netMock, len(busMocks))

	ctrl := gomock.NewController(t)
	noopBackgroundRouterMock := mock_types.NewMockRouter(ctrl)
	noopBackgroundRouterMock.EXPECT().Broadcast(gomock.Any()).Times(1)
	noopBackgroundRouterMock.EXPECT().Close().Times(len(busMocks))

	p2pModules = make(map[string]*p2pModule, len(busMocks))
	for i := range busMocks {
		host := netMock.Host(peerIDs[i])
		p2pMod, err := Create(
			busMocks[i],
			WithHost(host),
			// mock background router to prevent & ignore background message propagation.
			WithUnstakedActorRouter(noopBackgroundRouterMock),
		)
		require.NoError(t, err)
		p2pModules[validatorId(i+1)] = p2pMod.(*p2pModule)
	}
	return
}

func setupMockNetPeers(t *testing.T, netMock mocknet.Mocknet, numPeers int) (peerIDs []libp2pPeer.ID) {
	// Add a libp2p peers/hosts to the `MockNet` with the keypairs corresponding
	// to the genesis validators' keypairs
	for i, privKey := range keys[:numPeers] {
		peerInfo, err := utils.Libp2pAddrInfoFromPeer(&typesP2P.NetworkPeer{
			PublicKey:  privKey.PublicKey(),
			Address:    privKey.Address(),
			ServiceURL: validatorId(i + 1),
		})
		require.NoError(t, err)

		libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
		require.NoError(t, err)

		_, err = netMock.AddPeer(libp2pPrivKey, peerInfo.Addrs[0])
		require.NoError(t, err)

		peerIDs = append(peerIDs, peerInfo.ID)
	}

	// Link all peers such that any may dial/connect to any other.
	err := netMock.LinkAll()
	require.NoError(t, err)

	return peerIDs
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
		serviceURL := validatorId(i + 1)
		hostname, portStr, err := net.SplitHostPort(serviceURL)
		require.NoError(t, err)

		port, err := strconv.Atoi(portStr)
		require.NoError(t, err)

		cfg := &configs.Config{
			RootDirectory: "",
			PrivateKey:    valKeys[i].String(),
			P2P: &configs.P2PConfig{
				Hostname:       hostname,
				PrivateKey:     valKeys[i].String(),
				Port:           uint32(port),
				ConnectionType: types.ConnectionType_EmptyConnection,
				MaxNonces:      defaults.DefaultP2PMaxNonces,
			},
		}

		mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
		mockRuntimeMgr.EXPECT().GetConfig().Return(cfg).AnyTimes()
		mockRuntimeMgr.EXPECT().GetGenesis().Return(mockGenesisState).AnyTimes()
		mockRuntimeMgrs[i] = mockRuntimeMgr
	}
	return mockRuntimeMgrs
}

func createMockBuses(
	t *testing.T,
	runtimeMgrs []modules.RuntimeMgr,
	readWriteWaitGroup *sync.WaitGroup,
) []*mockModules.MockBus {
	mockBuses := make([]*mockModules.MockBus, len(runtimeMgrs))
	for i := range mockBuses {
		mockBuses[i] = createMockBus(t, runtimeMgrs[i], readWriteWaitGroup)
	}
	return mockBuses
}

func createMockBus(
	t *testing.T,
	runtimeMgr modules.RuntimeMgr,
	readWriteWaitGroup *sync.WaitGroup,
) *mockModules.MockBus {
	ctrl := gomock.NewController(t)
	mockBus := mockModules.NewMockBus(ctrl)
	mockBus.EXPECT().GetRuntimeMgr().Return(runtimeMgr).AnyTimes()
	modulesRegistry := runtime.NewModulesRegistry()
	mockBus.EXPECT().
		RegisterModule(gomock.Any()).
		DoAndReturn(func(m modules.Submodule) {
			modulesRegistry.RegisterModule(m)
			m.SetBus(mockBus)
		}).AnyTimes()
	mockBus.EXPECT().GetModulesRegistry().Return(modulesRegistry).AnyTimes()
	mockBus.EXPECT().PublishEventToBus(gomock.AssignableToTypeOf(&messaging.PocketEnvelope{})).
		Do(func(envelope *messaging.PocketEnvelope) {
			fmt.Println("[valId: unknown] Read")
			fmt.Printf("content type: %s\n", envelope.Content.GetTypeUrl())
			if readWriteWaitGroup != nil {
				readWriteWaitGroup.Done()
			}
		}).AnyTimes() // TECHDEBT: assert number of times. Consider `waitForEventsInternal` or similar as in consensus.
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
			ServiceUrl:      validatorId(i + 1),
			StakedAmount:    test_artifacts.DefaultStakeAmountString,
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

func prepareCurrentHeightProviderMock(t *testing.T, busMock *mockModules.MockBus) *mockModules.MockCurrentHeightProvider {
	ctrl := gomock.NewController(t)
	currentHeightProviderMock := mockModules.NewMockCurrentHeightProvider(ctrl)
	currentHeightProviderMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	currentHeightProviderMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	currentHeightProviderMock.EXPECT().SetBus(busMock).AnyTimes()
	currentHeightProviderMock.EXPECT().GetModuleName().
		Return(modules.CurrentHeightProviderSubmoduleName).
		AnyTimes()
	busMock.RegisterModule(currentHeightProviderMock)

	return currentHeightProviderMock
}

func preparePeerstoreProviderMock(
	t *testing.T,
	busMock *mockModules.MockBus,
	pstore typesP2P.Peerstore,
) *mock_types.MockPeerstoreProvider {
	ctrl := gomock.NewController(t)
	peerstoreProviderMock := mock_types.NewMockPeerstoreProvider(ctrl)
	peerstoreProviderMock.EXPECT().
		GetStakedPeerstoreAtHeight(gomock.Any()).
		Return(pstore, nil).
		AnyTimes()

	peerstoreProviderMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	peerstoreProviderMock.EXPECT().SetBus(busMock).AnyTimes()
	peerstoreProviderMock.EXPECT().GetModuleName().
		Return(peerstore_provider.PeerstoreProviderSubmoduleName).
		AnyTimes()

	return peerstoreProviderMock
}

// Persistence mock - only needed for validatorMap access
func preparePersistenceMock(t *testing.T, busMock *mockModules.MockBus, genesisState *genesis.GenesisState) *mockModules.MockPersistenceModule {
	ctrl := gomock.NewController(t)

	persistenceModuleMock := mockModules.NewMockPersistenceModule(ctrl)
	readCtxMock := mockModules.NewMockPersistenceReadContext(ctrl)

	readCtxMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	readCtxMock.EXPECT().GetAllStakedActors(gomock.Any()).DoAndReturn(func(height int64) ([]*coreTypes.Actor, error) {
		return testutil.Concatenate[*coreTypes.Actor](
			genesisState.GetValidators(),
			genesisState.GetServicers(),
			genesisState.GetFishermen(),
			genesisState.GetApplications(),
		), nil
	}).AnyTimes()
	persistenceModuleMock.EXPECT().NewReadContext(gomock.Any()).Return(readCtxMock, nil).AnyTimes()
	readCtxMock.EXPECT().Release().AnyTimes()

	persistenceModuleMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	persistenceModuleMock.EXPECT().SetBus(busMock).AnyTimes()
	persistenceModuleMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	busMock.RegisterModule(persistenceModuleMock)

	return persistenceModuleMock
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

	// TECHDEBT(#886): The number of times each telemetry event is expected
	// (below) is dependent on the number of redundant messages all validators see,
	// which is a function of the network size. Until this function is derived and
	// implemented, we cannot predict the number of times each event is expected.
	_ = expectedNumNetworkWrites

	eventMetricsAgentMock.EXPECT().EmitEvent(
		gomock.Any(),
		gomock.Any(),
		gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL),
		gomock.Any(),
	).Do(func(n, e any, l ...any) {
		log.Printf("[valId: %s] Write\n", valId)
		wg.Done()
	}).AnyTimes() // TECHDEBT: expect specific number of non-redundant writes once known.
	eventMetricsAgentMock.EXPECT().EmitEvent(
		gomock.Any(),
		gomock.Eq(telemetry.P2P_BROADCAST_MESSAGE_REDUNDANCY_PER_BLOCK_EVENT_METRIC_NAME),
		gomock.Any(),
		gomock.Any(), // nonce
		gomock.Any(),
		gomock.Any(), // blockHeight
	).Do(func(n, e any, l ...any) {
		log.Printf("[valId: %s] Write\n", valId)
		wg.Done()
	}).AnyTimes() // TECHDEBT: expect specific number of redundant writes once known.
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).AnyTimes()

	return eventMetricsAgentMock
}
