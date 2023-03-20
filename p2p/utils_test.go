package p2p

import (
	"context"
	"fmt"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"log"
	"net"
	"net/url"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/foxcpp/go-mockdns"
	"github.com/golang/mock/gomock"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	libp2pProtocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/p2p/protocol"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
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
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/telemetry"
)

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

const (
	serviceURLFormat  = "node%d.consensus:42069"
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

// CLEANUP: This could (should?) be a codebase-wide shared test helper
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
func createP2PModules(t *testing.T, busMocks []*mockModules.MockBus, hostMocks map[string]libp2pHost.Host) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(busMocks))
	for i := range busMocks {
		serviceURL := validatorId(i + 1)
		host := hostMocks[serviceURL]
		p2pMod, err := Create(busMocks[i], WithHostOption(host))
		require.NoError(t, err)
		p2pModules[serviceURL] = p2pMod.(*p2pModule)
	}
	return
}

// CONSIDERATION: this is one of a few places where we could de-duplicate test
// code if we had a conventional place to store packages intended for import
// only into tests.
func prepareDNSResolverMock(t *testing.T, serviceURLs []string) (done func()) {
	zones := make(map[string]mockdns.Zone)
	for i, u := range serviceURLs {
		// Perpend `scheme://` as serviceURLs are currently scheme-less.
		// Required to for parsing to produce useful results.
		// (see: https://pkg.go.dev/net/url@go1.20.2#URL)
		serviceURL, err := url.Parse(fmt.Sprintf("scheme://%s", u))
		require.NoError(t, err)

		ipStr := fmt.Sprintf("10.0.0.%d", i+1)

		if i >= 254 {
			panic(fmt.Sprintf("would generate invalid IPv4 address: %s", ipStr))
		}

		zones[fmt.Sprintf("%s.", serviceURL.Hostname())] = mockdns.Zone{
			A: []string{ipStr},
		}
	}

	srv, _ := mockdns.NewServerWithLogger(zones, noopLogger{}, false)
	srv.PatchNet(net.DefaultResolver)
	return func() {
		_ = srv.Close()
		mockdns.UnpatchNet(net.DefaultResolver)
	}
}

// NB: default logging behavior is too noisy.
// noopLogger implements go-mockdns's `mockdns.Logger` interface.
type noopLogger struct{}

func (nl noopLogger) Printf(format string, args ...interface{}) {
	// noop
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
				Hostname:       validatorId(i + 1),
				PrivateKey:     valKeys[i].String(),
				Port:           defaults.DefaultP2PPort,
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
	mockModulesRegistry.EXPECT().GetModule(peerstore_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(peerstore_provider.ModuleName)).AnyTimes()
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

// TODO _THIS_COMMIT: refactor this function
func prepareHostMock(t *testing.T, hostStreamHandlerMocks map[string]func(stream libp2pNetwork.Stream), pstore typesP2P.Peerstore, runtimeCfg *configs.Config, wg *sync.WaitGroup, expectedNumNetworkReads int) libp2pHost.Host {
	ctrl := gomock.NewController(t)

	privKey, err := cryptoPocket.NewPrivateKey(runtimeCfg.PrivateKey)
	require.NoError(t, err)

	peerInfo, err := utils.Libp2pAddrInfoFromPeer(&typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: runtimeCfg.P2P.Hostname,
	})
	require.NoError(t, err)

	hostMock := mock_types.NewMockHost(ctrl)
	hostMock.EXPECT().ID().Return(peerInfo.ID)

	selfLibp2pInfo, err := utils.Libp2pAddrInfoFromPeer(&typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: runtimeCfg.P2P.Hostname,
	})
	require.NoError(t, err)

	hostMock.EXPECT().Addrs().Return(selfLibp2pInfo.Addrs)
	hostMock.EXPECT().ID().Return(selfLibp2pInfo.ID)

	// TODO _THIS_COMMIT: move
	libp2pPStore, err := pstoremem.NewPeerstore()
	require.NoError(t, err)

	for _, peer := range pstore.GetPeerList() {
		peerInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
		require.NoError(t, err)

		var libp2pPubKey libp2pCrypto.PubKey
		if privKey.Address().Equals(peer.GetAddress()) {
			var libp2pPrivKey libp2pCrypto.PrivKey
			libp2pPrivKey, err = libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
			libp2pPStore.AddPrivKey(selfLibp2pInfo.ID, libp2pPrivKey)

			libp2pPubKey = libp2pPrivKey.GetPublic()
		} else {
			libp2pPubKey, err = utils.Libp2pPublicKeyFromPeer(peer)
		}
		require.NoError(t, err)

		libp2pPStore.AddAddrs(peerInfo.ID, peerInfo.Addrs, utils.DefaultPeerTTL)
		err = libp2pPStore.AddPubKey(peerInfo.ID, libp2pPubKey)
		require.NoError(t, err)
	}
	// --- END TODO

	networkMock := mock_types.NewMockNetwork(ctrl)
	networkMock.EXPECT().Notify(gomock.Any()).AnyTimes()
	networkMock.EXPECT().Peers().Return(libp2pPStore.Peers()).AnyTimes()
	networkMock.EXPECT().ConnsToPeer(gomock.Any()).Return(nil).AnyTimes()
	hostMock.EXPECT().Network().Return(networkMock).AnyTimes()

	hostMock.EXPECT().Peerstore().Return(libp2pPStore).AnyTimes()
	hostMock.EXPECT().Close().Return(nil).AnyTimes()
	hostMock.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, peerID libp2pPeer.ID, protocolIDs ...libp2pProtocol.ID) (libp2pNetwork.Stream, error) {
			eventsChannel := make(chan []byte, eventsChannelSize)

			handler, ok := hostStreamHandlerMocks[peerID.String()]
			require.Truef(t, ok, "no handle registered for peerID: %s", peerID)

			// Construct mock stream to return to sender caller.
			writeStreamMock := mock_types.NewMockStream(ctrl)
			writeStreamMock.EXPECT().Close().Return(nil).AnyTimes()
			writeStreamMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) (int, error) {
				eventsChannel <- data
				return len(data), nil
			}).AnyTimes()
			//}).Times(expectedNumNetworkReads)

			// Construct mock stream to pass to receiver stream handler func.
			readStreamMock := mock_types.NewMockStream(ctrl)
			readStreamMock.EXPECT().SetReadDeadline(gomock.Any()).Return(nil).AnyTimes()
			readStreamMock.EXPECT().Close().Return(nil).AnyTimes()
			readStreamMock.EXPECT().Read(gomock.Any()).DoAndReturn(func(data []byte) (int, error) {
				wg.Done()
				log.Printf("[valId: %s] Read\n", runtimeCfg.P2P.Hostname)
				data = <-eventsChannel
				return len(data), nil
			}).AnyTimes()
			//}).Times(expectedNumNetworkReads)

			selfLibp2pPubKey, err := libp2pCrypto.UnmarshalEd25519PublicKey(privKey.PublicKey().Bytes())
			require.NoError(t, err)

			connMock := mock_types.NewMockConn(ctrl)
			connMock.EXPECT().RemotePeer().Return(selfLibp2pInfo.ID).AnyTimes()
			connMock.EXPECT().RemotePublicKey().Return(selfLibp2pPubKey).AnyTimes()
			connMock.EXPECT().RemoteMultiaddr().Return(selfLibp2pInfo.Addrs[0]).AnyTimes()
			readStreamMock.EXPECT().Conn().Return(connMock).AnyTimes()

			// Call respective mock stream handler func
			handler(readStreamMock)

			return writeStreamMock, nil
		},
	).AnyTimes() // TODO _THIS_COMMIT: assert times

	// Collect stream handler functions as they're registered by peers.
	hostMock.EXPECT().SetStreamHandler(gomock.Eq(protocol.PoktProtocolID), gomock.Any()).
		DoAndReturn(func(protocolID libp2pProtocol.ID, handler func(stream libp2pNetwork.Stream)) {
			hostStreamHandlerMocks[peerInfo.ID.String()] = handler
		})

	// Gracefully ignore pubsub streams.
	hostMock.EXPECT().SetStreamHandler(gomock.Any(), gomock.Any()).Return().AnyTimes()
	return hostMock
}

// preparePStore populates a `Peerstore` instance of `*NetworkPeer`s derived
// from the passed runtime managers' configs' private keys.
func preparePStore(t *testing.T, runtimeMgrs []modules.RuntimeMgr) typesP2P.Peerstore {
	libp2pPeerInfos := make(map[string]libp2pPeer.AddrInfo)
	pstore := make(typesP2P.PeerAddrMap)
	for i, mgr := range runtimeMgrs {
		privKey, err := cryptoPocket.NewPrivateKey(mgr.GetConfig().PrivateKey)
		require.NoError(t, err)

		peer := &typesP2P.NetworkPeer{
			PublicKey: privKey.PublicKey(),
			Address:   privKey.Address(),
			// TECHDEBT: rename `validatorId`
			ServiceURL: validatorId(i + 1),
		}

		err = pstore.AddPeer(peer)
		require.NoError(t, err)

		peerInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
		require.NoError(t, err)

		libp2pPeerInfos[peer.GetAddress().String()] = peerInfo
	}
	return pstore
}
