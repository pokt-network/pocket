package p2p

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// TECHDEBT(#609): move & de-dup.
var testLocalServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)

func Test_Create_configureBootstrapNodes(t *testing.T) {
	defaultBootstrapNodes := strings.Split(defaults.DefaultP2PBootstrapNodesCsv, ",")
	privKey := cryptoPocket.GetPrivKeySeed(1)

	type args struct {
		initialBootstrapNodesCsv string
	}
	tests := []struct {
		name               string
		args               args
		wantBootstrapNodes []string
		wantErr            bool
	}{
		{
			name:               "unset boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args:               args{},
			wantErr:            false,
			wantBootstrapNodes: defaultBootstrapNodes,
		},
		{
			name: "empty string boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args: args{
				initialBootstrapNodesCsv: "",
			},
			wantErr:            false,
			wantBootstrapNodes: defaultBootstrapNodes,
		},
		{
			name: "untrimmed empty string boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args: args{
				initialBootstrapNodesCsv: "     ",
			},
			wantErr:            false,
			wantBootstrapNodes: defaultBootstrapNodes,
		},
		{
			name: "untrimmed string boostrap nodes should yield no error and return the trimmed urls",
			args: args{
				initialBootstrapNodesCsv: "     http://somenode:50832  ,  http://someothernode:50832  ",
			},
			wantErr:            false,
			wantBootstrapNodes: []string{"http://somenode:50832", "http://someothernode:50832"},
		},
		{
			name: "custom bootstrap nodes should yield no error and return the custom bootstrap node",
			args: args{
				initialBootstrapNodesCsv: "http://somenode:50832,http://someothernode:50832",
			},
			wantBootstrapNodes: []string{"http://somenode:50832", "http://someothernode:50832"},
			wantErr:            false,
		},
		{
			name: "malformed bootstrap nodes string should yield an error and return nil",
			args: args{
				initialBootstrapNodesCsv: "\n\n",
			},
			wantBootstrapNodes: []string(nil),
			wantErr:            true,
		},
		{
			name: "port number too high yields an error and empty list of bootstrap nodes",
			args: args{
				initialBootstrapNodesCsv: "http://somenode:99999",
			},
			wantBootstrapNodes: []string(nil),
			wantErr:            true,
		},
		{
			name: "negative port number yields an error and empty list of bootstrap nodes",
			args: args{
				initialBootstrapNodesCsv: "udp://somenode:-42069",
			},
			wantBootstrapNodes: []string(nil),
			wantErr:            true,
		},
		{
			name: "wrong protocol yields an error and empty list of bootstrap nodes",
			args: args{
				initialBootstrapNodesCsv: "udp://somenode:58884",
			},
			wantBootstrapNodes: []string(nil),
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
			mockBus := createMockBus(t, mockRuntimeMgr)

			genesisStateMock := createMockGenesisState(keys)
			persistenceMock := preparePersistenceMock(t, mockBus, genesisStateMock)
			mockBus.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()

			mockConsensusModule := mockModules.NewMockConsensusModule(ctrl)
			mockConsensusModule.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
			mockBus.EXPECT().GetConsensusModule().Return(mockConsensusModule).AnyTimes()
			mockRuntimeMgr.EXPECT().GetConfig().Return(&configs.Config{
				PrivateKey: privKey.String(),
				P2P: &configs.P2PConfig{
					BootstrapNodesCsv: tt.args.initialBootstrapNodesCsv,
					PrivateKey:        privKey.String(),
				},
			}).AnyTimes()
			mockBus.EXPECT().GetRuntimeMgr().Return(mockRuntimeMgr).AnyTimes()

			peer := &typesP2P.NetworkPeer{
				PublicKey:  privKey.PublicKey(),
				Address:    privKey.Address(),
				ServiceURL: testLocalServiceURL,
			}

			host := newLibp2pMockNetHost(t, privKey, peer)
			p2pMod, err := Create(mockBus, WithHostOption(host))
			if (err != nil) != tt.wantErr {
				t.Errorf("p2pModule.Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				actualBootstrapNodes := p2pMod.(*p2pModule).bootstrapNodes
				require.EqualValues(t, tt.wantBootstrapNodes, actualBootstrapNodes)
			}
		})
	}
}

func TestP2pModule_WithHostOption_Restart(t *testing.T) {
	ctrl := gomock.NewController(t)

	privKey := cryptoPocket.GetPrivKeySeed(1)
	mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
	mockBus := createMockBus(t, mockRuntimeMgr)

	genesisStateMock := createMockGenesisState(keys)
	persistenceMock := preparePersistenceMock(t, mockBus, genesisStateMock)
	mockBus.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()

	consensusModuleMock := mockModules.NewMockConsensusModule(ctrl)
	consensusModuleMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
	mockBus.EXPECT().GetConsensusModule().Return(consensusModuleMock).AnyTimes()

	telemetryModuleMock := baseTelemetryMock(t, nil)
	mockBus.EXPECT().GetTelemetryModule().Return(telemetryModuleMock).AnyTimes()

	mockRuntimeMgr.EXPECT().GetConfig().Return(&configs.Config{
		PrivateKey: privKey.String(),
		P2P: &configs.P2PConfig{
			PrivateKey: privKey.String(),
		},
	}).AnyTimes()
	mockBus.EXPECT().GetRuntimeMgr().Return(mockRuntimeMgr).AnyTimes()

	peer := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}

	mockNetHost := newLibp2pMockNetHost(t, privKey, peer)
	p2pMod, err := Create(mockBus, WithHostOption(mockNetHost))
	require.NoError(t, err)

	mod, ok := p2pMod.(*p2pModule)
	require.Truef(t, ok, "unknown module type: %T", mod)

	// start the module; should not create a new host
	err = mod.Start()
	require.NoError(t, err)

	// assert initial host matches the one provided via `WithHost`
	require.Equal(t, mockNetHost, mod.host, "initial hosts don't match")

	// stop the module; destroys module's host
	err = mod.Stop()
	require.NoError(t, err)

	// assert host matches still after restart
	err = mod.Start()
	require.NoError(t, err)
	require.Equal(t, mockNetHost, mod.host, "post-restart hosts don't match")
}

// TECHDEBT(#609): move & de-duplicate
func newLibp2pMockNetHost(t *testing.T, privKey cryptoPocket.PrivateKey, peer *typesP2P.NetworkPeer) libp2pHost.Host {
	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	libp2pMultiAddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.ServiceURL)
	require.NoError(t, err)

	libp2pMockNet := mocknet.New()
	host, err := libp2pMockNet.AddPeer(libp2pPrivKey, libp2pMultiAddr)
	require.NoError(t, err)

	return host
}

// TECHDEBT(#609): move & de-duplicate
func baseTelemetryMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := mockModules.NewMockTelemetryModule(ctrl)
	timeSeriesAgentMock := baseTelemetryTimeSeriesAgentMock(t)
	eventMetricsAgentMock := baseTelemetryEventMetricsAgentMock(t)

	telemetryMock.EXPECT().Start().Return(nil).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()

	return telemetryMock
}

// TECHDEBT(#609): move & de-duplicate
func baseTelemetryTimeSeriesAgentMock(t *testing.T) *mockModules.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeSeriesAgentMock := mockModules.NewMockTimeSeriesAgent(ctrl)
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeSeriesAgentMock
}

// TECHDEBT(#609): move & de-duplicate
func baseTelemetryEventMetricsAgentMock(t *testing.T) *mockModules.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mockModules.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}
