package p2p_test

import (
	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/runtime"
	"github.com/pokt-network/pocket/internal/testutil/telemetry"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/regen-network/gocuke"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	p2p_testutil "github.com/pokt-network/pocket/internal/testutil/p2p"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

const (
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
func createP2PModules(t *testing.T, busMocks []*mockModules.MockBus, netMock mocknet.Mocknet, serviceURLs []string) (p2pModules map[string]*p2pModule) {
	t.Helper()

	require.GreaterOrEqualf(t, len(serviceURLs), len(busMocks), "number of bus mocks must be less than or equal to the number of service URLs")

	peerIDs := p2p_testutil.SetupMockNetPeers(t, netMock, keys[:len(busMocks)], serviceURLs)
	p2pModules = make(map[string]*p2pModule, len(busMocks))
	for i := range busMocks {
		host := netMock.Host(peerIDs[i])
		p2pMod, err := Create(busMocks[i], WithHostOption(host))
		require.NoError(t, err)
		p2pModules[serviceURLs[i]] = p2pMod.(*p2pModule)
	}
	return
}

// createMockRuntimeMgrs creates `numValidators` instances of mocked `RuntimeMgr` that are essentially
// representing the runtime environments of the validators that we will use in our tests
func createMockRuntimeMgrs(t *testing.T, numValidators int) []modules.RuntimeMgr {
	mockRuntimeMgrs := make([]modules.RuntimeMgr, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys, keys[:numValidators])
	mockGenesisState := runtime_testutil.GenesisWithSequentialServiceURLs(t, valKeys)
	for i := range mockRuntimeMgrs {
		mockRuntimeMgrs[i] = runtime_testutil.BaseRuntimeManagerMock(
			t, valKeys[i],
			p2p_testutil.NewServiceURL(i+1),
			mockGenesisState,
		)
	}
	return mockRuntimeMgrs
}

func createMockBuses(t *testing.T, runtimeMgrs []modules.RuntimeMgr, wg *sync.WaitGroup) []*mockModules.MockBus {
	mockBuses := make([]*mockModules.MockBus, len(runtimeMgrs))
	for i := range mockBuses {
		handlerFactory := func(t gocuke.TestingT, bus modules.Bus) testutil.BusEventHandler {
			return func(data *messaging.PocketEnvelope) {
				wg.Done()
			}
		}
		mockBuses[i] = testutil.BusMockWithEventHandler(t, runtimeMgrs[i], handlerFactory)
	}
	return mockBuses
}

// TODO_THIS_COMMIT: refactor
// Telemetry mock - Needed to help with proper counts for number of expected network writes
func prepareTelemetryMock(t *testing.T, busMock *mockModules.MockBus, valId string, wg *sync.WaitGroup, expectedNumNetworkWrites int) *mockModules.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := mockModules.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := telemetry_testutil.BaseTimeSeriesAgentMock(t)
	// TODO_THIS_COMMIT: refactor
	eventMetricsAgentMock := telemetry_testutil.PrepareEventMetricsAgentMock(t, valId, wg, expectedNumNetworkWrites)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()

	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()
	telemetryMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	telemetryMock.EXPECT().SetBus(busMock).AnyTimes()
	busMock.RegisterModule(telemetryMock)

	return telemetryMock
}
