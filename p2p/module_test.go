package p2p

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

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

			genesisStateMock := createMockGenesisState(keys[:1])
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
				ServiceURL: "10.0.0.1:42069",
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

// TECHDEBT: de-duplicate
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
