package p2p

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/stretchr/testify/require"
)

func Test_Create_configureBootstrapNodes(t *testing.T) {
	defaultBootstrapNodes := strings.Split(defaults.DefaultP2PBootstrapNodesCsv, ",")
	key := generateKey(t, 1)

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
				initialBootstrapNodesCsv: "udp://somenode:-8080",
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
			mockConsensusModule := mockModules.NewMockConsensusModule(ctrl)
			mockBus.EXPECT().GetConsensusModule().Return(mockConsensusModule).AnyTimes()
			mockRuntimeMgr.EXPECT().GetConfig().Return(&configs.Config{
				PrivateKey: key.String(),
				P2P: &configs.P2PConfig{
					BootstrapNodesCsv: tt.args.initialBootstrapNodesCsv,
					PrivateKey:        key.String(),
				},
			}).AnyTimes()
			mockBus.EXPECT().GetRuntimeMgr().Return(mockRuntimeMgr).AnyTimes()

			var p2pMod modules.Module
			p2pMod, err := Create(mockBus)
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
