//go:build test

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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
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
			libp2pMockNet := mocknet.New()

			ctrl := gomock.NewController(t)
			mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
			mockBus := createMockBus(t, mockRuntimeMgr, nil)

			genesisStateMock := createMockGenesisState(keys)
			persistenceMock := preparePersistenceMock(t, mockBus, genesisStateMock)
			mockBus.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()

			mockConsensusModule := mockModules.NewMockConsensusModule(ctrl)
			mockConsensusModule.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
			mockBus.EXPECT().GetConsensusModule().Return(mockConsensusModule).AnyTimes()

			currentHeightProviderMock := prepareCurrentHeightProviderMock(t, mockBus)
			mockBus.RegisterModule(currentHeightProviderMock)

			pstore := new(typesP2P.PeerAddrMap)
			pstoreProviderMock := preparePeerstoreProviderMock(t, mockBus, pstore)
			mockBus.RegisterModule(pstoreProviderMock)

			mockRuntimeMgr.EXPECT().GetConfig().Return(&configs.Config{
				PrivateKey: privKey.String(),
				P2P: &configs.P2PConfig{
					BootstrapNodesCsv: tt.args.initialBootstrapNodesCsv,
					PrivateKey:        privKey.String(),
					MaxNonces:         100,
				},
			}).AnyTimes()
			mockBus.EXPECT().GetRuntimeMgr().Return(mockRuntimeMgr).AnyTimes()

			peer := &typesP2P.NetworkPeer{
				PublicKey:  privKey.PublicKey(),
				Address:    privKey.Address(),
				ServiceURL: testLocalServiceURL,
			}

			host := newMockNetHost(t, libp2pMockNet, privKey, peer)
			p2pMod, err := Create(mockBus, WithHost(host))
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
	privKey := cryptoPocket.GetPrivKeySeed(1)

	peer := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}

	libp2pMockNet := mocknet.New()
	host := newMockNetHost(t, libp2pMockNet, privKey, peer)

	mod := newP2PModule(t, privKey, WithHost(host))

	// start the module; should not create a new host
	err := mod.Start()
	require.NoError(t, err)

	// assert initial host matches the one provided via `WithHost`
	require.Equal(t, host, mod.host, "initial hosts don't match")

	// stop the module; destroys module's host
	err = mod.Stop()
	require.NoError(t, err)

	// assert host does *not* match after restart
	err = mod.Start()
	require.NoError(t, err)
	require.NotEqual(t, host, mod.host, "post-restart hosts don't match")
}

func TestP2pModule_InvalidNonce(t *testing.T) {
	privKey := cryptoPocket.GetPrivKeySeed(1)

	peer := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}

	libp2pMockNet := mocknet.New()
	host := newMockNetHost(t, libp2pMockNet, privKey, peer)

	mod := newP2PModule(
		t, privKey,
		WithHost(host),
	)
	err := mod.Start()
	require.NoError(t, err)

	// Use zero value nonce
	poktEnvelope := &messaging.PocketEnvelope{
		Content: &anypb.Any{},
	}
	poktEnvelopeBz, err := proto.Marshal(poktEnvelope)
	require.NoError(t, err)

	err = mod.handlePocketEnvelope(poktEnvelopeBz)
	require.ErrorIs(t, err, typesP2P.ErrInvalidNonce)

	// Explicitly set the nonce to 0
	poktEnvelope = &messaging.PocketEnvelope{
		Content: &anypb.Any{},
		// 0 should be an invalid nonce value
		Nonce: 0,
	}
	poktEnvelopeBz, err = proto.Marshal(poktEnvelope)
	require.NoError(t, err)

	err = mod.handlePocketEnvelope(poktEnvelopeBz)
	require.ErrorIs(t, err, typesP2P.ErrInvalidNonce)
}

// TECHDEBT(#609): move & de-duplicate
func newP2PModule(t *testing.T, privKey cryptoPocket.PrivateKey, opts ...modules.ModuleOption) *p2pModule {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
	mockBus := createMockBus(t, mockRuntimeMgr, nil)

	genesisStateMock := createMockGenesisState(nil)
	persistenceMock := preparePersistenceMock(t, mockBus, genesisStateMock)
	mockBus.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()

	consensusModuleMock := mockModules.NewMockConsensusModule(ctrl)
	consensusModuleMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
	mockBus.EXPECT().GetConsensusModule().Return(consensusModuleMock).AnyTimes()

	currentHeightProviderMock := prepareCurrentHeightProviderMock(t, mockBus)
	mockBus.RegisterModule(currentHeightProviderMock)
	mockBus.EXPECT().
		GetCurrentHeightProvider().
		Return(currentHeightProviderMock).
		AnyTimes()

	pstore := new(typesP2P.PeerAddrMap)
	pstoreProviderMock := preparePeerstoreProviderMock(t, mockBus, pstore)
	mockBus.RegisterModule(pstoreProviderMock)

	telemetryModuleMock := baseTelemetryMock(t, nil)
	mockBus.EXPECT().GetTelemetryModule().Return(telemetryModuleMock).AnyTimes()

	mockRuntimeMgr.EXPECT().GetConfig().Return(&configs.Config{
		PrivateKey: privKey.String(),
		P2P: &configs.P2PConfig{
			PrivateKey: privKey.String(),
			MaxNonces:  defaults.DefaultP2PMaxNonces,
		},
	}).AnyTimes()
	mockBus.EXPECT().GetRuntimeMgr().Return(mockRuntimeMgr).AnyTimes()
	p2pMod, err := Create(mockBus, opts...)
	require.NoError(t, err)

	mod, ok := p2pMod.(*p2pModule)
	require.Truef(t, ok, "unknown module type: %T", mod)

	return mod
}

// TECHDEBT(#609): move & de-duplicate
func newMockNetHost(
	t *testing.T,
	libp2pMockNet mocknet.Mocknet,
	privKey cryptoPocket.PrivateKey,
	peer *typesP2P.NetworkPeer,
) libp2pHost.Host {
	t.Helper()

	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	libp2pMultiAddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.ServiceURL)
	require.NoError(t, err)

	host, err := libp2pMockNet.AddPeer(libp2pPrivKey, libp2pMultiAddr)
	require.NoError(t, err)

	return host
}

// TECHDEBT(#609): move & de-duplicate
func baseTelemetryMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockTelemetryModule {
	t.Helper()

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
	t.Helper()

	ctrl := gomock.NewController(t)
	timeSeriesAgentMock := mockModules.NewMockTimeSeriesAgent(ctrl)
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeSeriesAgentMock
}

// TECHDEBT(#609): move & de-duplicate
func baseTelemetryEventMetricsAgentMock(t *testing.T) *mockModules.MockEventMetricsAgent {
	t.Helper()

	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mockModules.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}
