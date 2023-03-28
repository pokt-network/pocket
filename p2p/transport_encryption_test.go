package p2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/p2p/protocol"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

func TestP2pModule_Insecure_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	hostname := "127.0.0.1"

	privKey := cryptoPocket.GetPrivKeySeed(1)

	mockConsensusModule := mockModules.NewMockConsensusModule(ctrl)
	mockConsensusModule.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(&configs.Config{
		PrivateKey: privKey.String(),
		P2P: &configs.P2PConfig{
			PrivateKey:     privKey.String(),
			Hostname:       hostname,
			Port:           defaults.DefaultP2PPort,
			ConnectionType: types.ConnectionType_TCPConnection,
		},
	}).AnyTimes()

	timeSeriesAgentMock := prepareNoopTimeSeriesAgentMock(t)
	eventMetricsAgentMock := mockModules.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	telemetryMock := mockModules.NewMockTelemetryModule(ctrl)
	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()

	busMock := createMockBus(t, runtimeMgrMock)
	busMock.EXPECT().GetConsensusModule().Return(mockConsensusModule).AnyTimes()
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	genesisStateMock := createMockGenesisState(keys[:1])
	persistenceMock := preparePersistenceMock(t, busMock, genesisStateMock)
	busMock.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()

	telemetryMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	telemetryMock.EXPECT().SetBus(busMock).AnyTimes()

	p2pMod, err := Create(busMock)
	require.NoError(t, err)

	err = p2pMod.Start()
	require.NoError(t, err)
	defer func() {
		_ = p2pMod.Stop()
	}()

	// Setup cleartext transport node
	clearNodeMultiAddrStr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", defaults.DefaultP2PPort+1)
	clearNodeAddr, err := multiaddr.NewMultiaddr(clearNodeMultiAddrStr)
	require.NoError(t, err)

	clearNode, err := libp2p.New(libp2p.NoSecurity, libp2p.ListenAddrs(clearNodeAddr))
	require.NoError(t, err)
	defer func() {
		_ = clearNode.Close()
	}()

	p2pModPeer := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: fmt.Sprintf("%s:%d", hostname, defaults.DefaultP2PPort),
	}

	libp2pPeerInfo, err := utils.Libp2pAddrInfoFromPeer(p2pModPeer)
	require.NoError(t, err)

	libp2pPubKey, err := utils.Libp2pPublicKeyFromPeer(p2pModPeer)
	require.NoError(t, err)

	clearNode.Peerstore().AddAddrs(libp2pPeerInfo.ID, libp2pPeerInfo.Addrs, time.Hour)
	err = clearNode.Peerstore().AddPubKey(libp2pPeerInfo.ID, libp2pPubKey)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = clearNode.NewStream(ctx, libp2pPeerInfo.ID, protocol.PoktProtocolID)
	require.ErrorContains(t, err, "failed to negotiate security protocol: protocols not supported:")
}
