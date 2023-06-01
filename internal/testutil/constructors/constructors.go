package constructors

import (
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	runtime_testutil "github.com/pokt-network/pocket/internal/testutil/runtime"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

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
func NewP2PModulesAndMocknet(
	t gocuke.TestingT,
	count int,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
) (
	buses map[serviceURLStr]*mock_modules.MockBus,
	p2pModules map[serviceURLStr]modules.P2PModule,
	libp2pNetworkMock mocknet.Mocknet,
) {
	libp2pNetworkMock = mocknet.New()
	// destroy mocknet on test cleanup
	t.Cleanup(func() {
		err := libp2pNetworkMock.Close()
		require.NoError(t, err)
	})

	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// ID collisions
	privKeys := testutil.LoadLocalnetPrivateKeys(t, count)

	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// serviceURL collisions
	serviceURLs := p2p_testutil.SequentialServiceURLs(t, count)
	_ = p2p_testutil.SetupMockNetPeers(t, libp2pNetworkMock, privKeys, serviceURLs)

	for i, serviceURL := range serviceURLs {
		// TECHDEBT: refactor
		t.Logf("hosts: %v", libp2pNetworkMock.Hosts())
		host := libp2pNetworkMock.Hosts()[i]
		buses[serviceURL], p2pModules[serviceURL] = NewP2PModuleWithHost(
			t, privKeys[i],
			serviceURLs[i],
			host,
			genesisState,
			busEventHandlerFactory,
		)
	}
	return buses, p2pModules, libp2pNetworkMock
}

// TODO_THIS_COMMIT: consider following create factory convention (?)
func NewP2PModuleWithHost(
	t gocuke.TestingT,
	privKey cryptoPocket.PrivateKey,
	serviceURL string,
	host libp2pHost.Host,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
) (*mock_modules.MockBus, modules.P2PModule) {
	t.Helper()

	runtimeMgrMock := runtime_testutil.BaseRuntimeManagerMock(t, privKey, serviceURL, genesisState)
	busMock := testutil.BusMockWithEventHandler(t, runtimeMgrMock, busEventHandlerFactory)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()

	mod, err := p2p.Create(busMock, p2p.WithHostOption(host))
	require.NoError(t, err)

	return busMock, mod.(modules.P2PModule)
}
