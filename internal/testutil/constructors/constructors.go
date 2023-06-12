package constructors

import (
	"github.com/foxcpp/go-mockdns"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/pokt-network/pocket/internal/testutil/bus"
	consensus_testutil "github.com/pokt-network/pocket/internal/testutil/consensus"
	persistence_testutil "github.com/pokt-network/pocket/internal/testutil/persistence"
	telemetry_testutil "github.com/pokt-network/pocket/internal/testutil/telemetry"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"net"
	"strconv"

	"github.com/pokt-network/pocket/internal/testutil"
	p2p_testutil "github.com/pokt-network/pocket/internal/testutil/p2p"
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
	dnsSrv *mockdns.Server,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
	notifiee libp2pNetwork.Notifiee,
) (
	buses map[serviceURLStr]*mock_modules.MockBus,
	libp2pNetworkMock mocknet.Mocknet,
	p2pModules map[serviceURLStr]modules.P2PModule,
) {
	libp2pNetworkMock = p2p_testutil.NewLibp2pNetworkMock(t)
	serviceURLKeyMap := testutil.SequentialServiceURLPrivKeyMap(t, count)

	buses, p2pModules = NewBusesAndP2PModules(
		t, busEventHandlerFactory,
		dnsSrv,
		genesisState,
		libp2pNetworkMock,
		serviceURLKeyMap,
		notifiee,
	)
	err := libp2pNetworkMock.LinkAll()
	require.NoError(t, err)

	return buses, libp2pNetworkMock, p2pModules
}

// TODO_THIS_COMMIT: rename / move, if possible
func NewP2PModule(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
	dnsSrv *mockdns.Server,
	libp2pNetworkMock mocknet.Mocknet,
	notifiee libp2pNetwork.Notifiee,
	// TODO_THIS_COMMIT: consider *p2p.P2PModule instead
) modules.P2PModule {
	genesisState := busMock.GetRuntimeMgr().GetGenesis()

	_ = consensus_testutil.BaseConsensusMock(t, busMock)
	_ = persistence_testutil.BasePersistenceMock(t, busMock, genesisState)

	// -- option 1
	_ = telemetry_testutil.BaseTelemetryMock(t, busMock)

	// -- option 2
	//_ = telemetry_testutil.WithTimeSeriesAgent(
	//	t, telemetry_testutil.MinimalTelemetryMock(t, busMock),
	//)

	p2pCfg := busMock.GetRuntimeMgr().GetConfig().P2P
	serviceURL := net.JoinHostPort(p2pCfg.Hostname, strconv.Itoa(int(p2pCfg.Port)))
	privKey, err := cryptoPocket.NewPrivateKey(p2pCfg.PrivateKey)
	require.NoError(t, err)

	// MUST register DNS before instantiating P2PModule
	testutil.AddServiceURLZone(t, dnsSrv, serviceURL)

	host := testutil.NewMocknetHost(t, libp2pNetworkMock, privKey, notifiee)
	return NewP2PModuleWithHost(t, busMock, host)
}

// TODO_THIS_TEST: need this?
func NewBusesAndP2PModules(
	t gocuke.TestingT,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
	dnsSrv *mockdns.Server,
	genesisState *genesis.GenesisState,
	libp2pNetworkMock mocknet.Mocknet,
	serviceURLKeyMap map[serviceURLStr]cryptoPocket.PrivateKey,
	notifiee libp2pNetwork.Notifiee,
) (
	busMocks map[serviceURLStr]*mock_modules.MockBus,
	// TODO_THIS_COMMIT: consider *p2p.P2PModule instead
	p2pModules map[serviceURLStr]modules.P2PModule,
) {
	busMocks = make(map[serviceURLStr]*mock_modules.MockBus)
	p2pModules = make(map[serviceURLStr]modules.P2PModule)

	for serviceURL, privKey := range serviceURLKeyMap {
		busMock := bus_testutil.NewBus(
			t, privKey,
			serviceURL,
			genesisState,
			busEventHandlerFactory,
		)
		busMocks[serviceURL] = busMock

		p2pModules[serviceURL] = NewP2PModule(
			t, busMock,
			dnsSrv,
			libp2pNetworkMock,
			notifiee,
		)
	}
	return busMocks, p2pModules
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

	busMock := bus_testutil.NewBus(t, privKey, serviceURL, genesisState, busEventHandlerFactory)
	return busMock, NewP2PModuleWithHost(t, busMock, host)
}

// TODO_THIS_COMMIT: rename; consider returning *p2p.P2PModule instead
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
