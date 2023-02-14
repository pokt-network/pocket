package p2p

import (
	"strings"
	"testing"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/stretchr/testify/require"
)

func Test_configureBootstrapNodes(t *testing.T) {
	defaultBootstrapNodes := strings.Split(defaults.DefaultP2PBootstrapNodesCsv, ",")
	type args struct {
		p2pCfg *configs.P2PConfig
		m      *p2pModule
	}
	tests := []struct {
		name               string
		args               args
		wantBootstrapNodes []string
		wantErr            bool
	}{
		{
			name: "unset boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args: args{
				p2pCfg: &configs.P2PConfig{},
				m:      new(p2pModule),
			},
			wantErr:            false,
			wantBootstrapNodes: defaultBootstrapNodes,
		},
		{
			name: "empty string boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args: args{
				p2pCfg: &configs.P2PConfig{
					BootstrapNodesCsv: "",
				},
				m: new(p2pModule),
			},
			wantErr:            false,
			wantBootstrapNodes: defaultBootstrapNodes,
		},
		{
			name: "untrimmed empty string boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args: args{
				p2pCfg: &configs.P2PConfig{
					BootstrapNodesCsv: "     ",
				},
				m: new(p2pModule),
			},
			wantErr:            false,
			wantBootstrapNodes: defaultBootstrapNodes,
		},
		{
			name: "untrimmed string boostrap nodes should yield no error and return DefaultP2PBootstrapNodes",
			args: args{
				p2pCfg: &configs.P2PConfig{
					BootstrapNodesCsv: "     http://somenode:50832  ",
				},
				m: new(p2pModule),
			},
			wantErr:            false,
			wantBootstrapNodes: []string{"http://somenode:50832"},
		},
		{
			name: "custom bootstrap nodes should yield no error and return the custom bootstrap node",
			args: args{
				p2pCfg: &configs.P2PConfig{
					BootstrapNodesCsv: "http://somenode:50832,http://someothernode:50832",
				},
				m: new(p2pModule),
			},
			wantBootstrapNodes: []string{"http://somenode:50832", "http://someothernode:50832"},
			wantErr:            false,
		},
		{
			name: "malformed bootstrap nodes string should yield an error and return nil",
			args: args{
				p2pCfg: &configs.P2PConfig{
					BootstrapNodesCsv: "\n\n",
				},
				m: &p2pModule{},
			},
			wantBootstrapNodes: []string(nil),
			wantErr:            true,
		},
		{
			name: "invalid hostname:port pattern for bootstrap nodes string should yield an error and return nil",
			args: args{
				p2pCfg: &configs.P2PConfig{
					BootstrapNodesCsv: "http://somenode:99999",
				},
				m: &p2pModule{},
			},
			wantBootstrapNodes: []string(nil),
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := configureBootstrapNodes(tt.args.p2pCfg, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("configureBootstrapNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.EqualValues(t, tt.wantBootstrapNodes, tt.args.m.bootstrapNodes)
		})
	}
}
