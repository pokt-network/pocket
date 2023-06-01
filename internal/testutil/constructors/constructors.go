package constructors

import (
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	consensus_testutil "github.com/pokt-network/pocket/internal/testutil/consensus"
	persistence_testutil "github.com/pokt-network/pocket/internal/testutil/persistence"
	telemetry_testutil "github.com/pokt-network/pocket/internal/testutil/telemetry"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/internal/testutil"
	p2p_testutil "github.com/pokt-network/pocket/internal/testutil/p2p"
	runtime_testutil "github.com/pokt-network/pocket/internal/testutil/runtime"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/runtime/genesis"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
)

type serviceURLStr = string

// NewP2PMocknetModules returns a map of peer IDs to P2PModules using libp2p mocknet hosts.
func NewBusesMocknetAndP2PModules(
	t gocuke.TestingT,
	count int,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
) (
	buses map[serviceURLStr]*mock_modules.MockBus,
	libp2pNetworkMock mocknet.Mocknet,
	p2pModules map[serviceURLStr]modules.P2PModule,
) {
	// TODO_THIS_COMMIT: refactor
	dnsSrv := testutil.MinimalDNSMock(t)

	libp2pNetworkMock = mocknet.New()
	// destroy mocknet on test cleanup
	t.Cleanup(func() {
		err := libp2pNetworkMock.Close()
		require.NoError(t, err)
	})

	buses = make(map[serviceURLStr]*mock_modules.MockBus)
	p2pModules = make(map[serviceURLStr]modules.P2PModule)
	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// ID collisions
	privKeys := testutil.LoadLocalnetPrivateKeys(t, count)
	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// serviceURL collisions
	serviceURLs := p2p_testutil.SequentialServiceURLs(t, count)
	for i, serviceURL := range serviceURLs {
		if len(privKeys) <= i {
			t.Logf("WARNING: not enough private keys for %d service URLs", len(serviceURLs))
			break
		}

		privKey := privKeys[i]
		busMock := NewBus(t, privKey, serviceURL, genesisState, busEventHandlerFactory)
		buses[serviceURL] = busMock

		// TODO_THIS_COMMIT: refactor
		_ = consensus_testutil.BaseConsensusMock(t, busMock)
		_ = persistence_testutil.BasePersistenceMock(t, busMock, genesisState)
		//_ = telemetry_testutil.BaseTelemetryMock(t, busMock)
		_ = telemetry_testutil.WithTimeSeriesAgent(
			t, telemetry_testutil.MinimalTelemetryMock(t, busMock),
		)

		// MUST register DNS before instantiating P2PModule
		testutil.AddServiceURLZone(t, dnsSrv, serviceURL)

		host := p2p_testutil.NewMocknetHost(t, libp2pNetworkMock, privKey)
		p2pModules[serviceURL] = NewP2PModuleWithHost(t, busMock, host)
	}
	err := libp2pNetworkMock.LinkAll()
	require.NoError(t, err)

	return buses, libp2pNetworkMock, p2pModules
}

// TODO_THIS_TEST: need this?
func NewP2PModules(
	t gocuke.TestingT,
	privKeys []cryptoPocket.PrivateKey,
	busMock *mock_modules.MockBus,
	libp2pNetworkMock mocknet.Mocknet,
) (
	p2pModules map[serviceURLStr]modules.P2PModule,
) {
	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// serviceURL collisions
	serviceURLs := p2p_testutil.SequentialServiceURLs(t, len(privKeys))
	_ = p2p_testutil.SetupMockNetPeers(t, libp2pNetworkMock, privKeys, serviceURLs)

	for i, serviceURL := range serviceURLs {
		host := libp2pNetworkMock.Hosts()[i]
		// TECHDEBT: refactor
		p2pModules[serviceURL] = NewP2PModuleWithHost(t, busMock, host)
	}
	return p2pModules
}

// TODO_THIS_TEST: need this?
// TODO_THIS_COMMIT: consider following create factory convention (?)
func NewBusesAndP2PModuleWithHost(
	t gocuke.TestingT,
	privKey cryptoPocket.PrivateKey,
	serviceURL string,
	host libp2pHost.Host,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
) (*mock_modules.MockBus, modules.P2PModule) {
	t.Helper()

	busMock := NewBus(t, privKey, serviceURL, genesisState, busEventHandlerFactory)
	return busMock, NewP2PModuleWithHost(t, busMock, host)
}

func NewBus(
	t gocuke.TestingT,
	privKey cryptoPocket.PrivateKey,
	serviceURL string,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
) *mock_modules.MockBus {
	t.Helper()

	runtimeMgrMock := runtime_testutil.BaseRuntimeManagerMock(
		t, privKey,
		serviceURL,
		genesisState,
	)
	busMock := testutil.BusMockWithEventHandler(t, runtimeMgrMock, busEventHandlerFactory)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	return busMock
}

func NewP2PModuleWithHost(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
	host libp2pHost.Host,
) modules.P2PModule {
	t.Helper()

	mod, err := p2p.Create(busMock, p2p.WithHostOption(host))
	require.NoError(t, err)

	return mod.(modules.P2PModule)
}
