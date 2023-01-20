package runtime

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var expectedGenesis = &genesis.GenesisState{
	GenesisTime: &timestamppb.Timestamp{
		Seconds: 1663610702,
		Nanos:   405401000,
	},
	ChainId:       "testnet",
	MaxBlockBytes: 4000000,
	Pools: []*types.Account{
		{
			Address: "DAO",
			Amount:  "100000000000000",
		},
		{
			Address: "FeeCollector",
			Amount:  "0",
		},
		{
			Address: "AppStakePool",
			Amount:  "100000000000000",
		},
		{
			Address: "ValidatorStakePool",
			Amount:  "100000000000000",
		},
		{
			Address: "ServiceNodeStakePool",
			Amount:  "100000000000000",
		},
		{
			Address: "FishermanStakePool",
			Amount:  "100000000000000",
		},
	},
	Accounts: []*types.Account{
		{
			Address: "004aaa1375c252eecc1e8737b2062b967532b07f",
			Amount:  "100000000000000",
		},
		{
			Address: "003aabccc8c3938c772e783ffdd9b9fabdc118cd",
			Amount:  "100000000000000",
		},
		{
			Address: "002aad32c4e0924b1182370cddb0b3071bb84a86",
			Amount:  "100000000000000",
		},
		{
			Address: "001aaba8fa5409efe6cd881649743bc7576acd46",
			Amount:  "100000000000000",
		},
		{
			Address: "43d9ea9d9ad9c58bb96ec41340f83cb2cabb6496",
			Amount:  "100000000000000",
		},
		{
			Address: "9ba047197ec043665ad3f81278ab1f5d3eaf6b8b",
			Amount:  "100000000000000",
		},
		{
			Address: "88a792b7aca673620132ef01f50e62caa58eca83",
			Amount:  "100000000000000",
		},
	},
	Applications: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_APP,
			Address:         "88a792b7aca673620132ef01f50e62caa58eca83",
			PublicKey:       "5f78658599943dc3e623539ce0b3c9fe4e192034a1e3fef308bc9f96915754e0",
			Chains:          []string{"0001"},
			GenericParam:    "1000000",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "88a792b7aca673620132ef01f50e62caa58eca83",
		},
	},
	Validators: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "001aaba8fa5409efe6cd881649743bc7576acd46",
			PublicKey:       "f48654c9bffccd7a858dc5577551ff650f8df9f1ec5bb668f339f594f2380ba1",
			Chains:          nil,
			GenericParam:    "node1.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "001aaba8fa5409efe6cd881649743bc7576acd46",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "002aad32c4e0924b1182370cddb0b3071bb84a86",
			PublicKey:       "caa495ca5958ff1ea9361716da270f5a03ca7e9cb85f955393e97264880d2c80",
			Chains:          nil,
			GenericParam:    "node2.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "002aad32c4e0924b1182370cddb0b3071bb84a86",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "003aabccc8c3938c772e783ffdd9b9fabdc118cd",
			PublicKey:       "130584fbf284bf68010b643a868b89dbbee68dc72d4e8f5e6c9bb9b48df67cd4",
			Chains:          nil,
			GenericParam:    "node3.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "003aabccc8c3938c772e783ffdd9b9fabdc118cd",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "004aaa1375c252eecc1e8737b2062b967532b07f",
			PublicKey:       "f511f0037512e802a584a1ef714790013f3db8d79e5f62cc2cae6902e1d7410b",
			Chains:          nil,
			GenericParam:    "node4.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "004aaa1375c252eecc1e8737b2062b967532b07f",
		},
	},
	ServiceNodes: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_SERVICENODE,
			Address:         "43d9ea9d9ad9c58bb96ec41340f83cb2cabb6496",
			PublicKey:       "16cd0a304c38d76271f74dd3c90325144425d904ef1b9a6fbab9b201d75a998b",
			Chains:          []string{"0001"},
			GenericParam:    "node1.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "43d9ea9d9ad9c58bb96ec41340f83cb2cabb6496",
		},
	},
	Fishermen: []*types.Actor{
		{
			ActorType:       types.ActorType_ACTOR_TYPE_FISH,
			Address:         "9ba047197ec043665ad3f81278ab1f5d3eaf6b8b",
			PublicKey:       "68efd26af01692fcd77dc135ca1de69ede464e8243e6832bd6c37f282db8c9cb",
			Chains:          []string{"0001"},
			GenericParam:    "node1.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "9ba047197ec043665ad3f81278ab1f5d3eaf6b8b",
		},
	},
	Params: test_artifacts.DefaultParams(),
}

func TestNewManagerFromReaders(t *testing.T) {
	type args struct {
		configReader  io.Reader
		genesisReader io.Reader
		options       []func(*Manager)
	}

	buildConfigBytes, err := os.ReadFile("../build/config/config1.json")
	if err != nil {
		require.NoError(t, err)
	}

	buildGenesisBytes, err := os.ReadFile("../build/config/genesis.json")
	if err != nil {
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		args      args
		want      *Manager
		assertion require.ComparisonAssertionFunc
	}{
		{
			name: "reading from the build directory",
			args: args{
				configReader:  strings.NewReader(string(buildConfigBytes)),
				genesisReader: strings.NewReader(string(buildGenesisBytes)),
			},
			want: &Manager{
				config: &configs.Config{
					RootDirectory: "/go/src/github.com/pocket-network",
					PrivateKey:    "7e1cd00241f5a743dc1a42919ef56b0d32d8e868eae216c03d44c52faa415ef2f48654c9bffccd7a858dc5577551ff650f8df9f1ec5bb668f339f594f2380ba1",
					Consensus: &configs.ConsensusConfig{
						PrivateKey:      "7e1cd00241f5a743dc1a42919ef56b0d32d8e868eae216c03d44c52faa415ef2f48654c9bffccd7a858dc5577551ff650f8df9f1ec5bb668f339f594f2380ba1",
						MaxMempoolBytes: 500000000,
						PacemakerConfig: &configs.PacemakerConfig{
							TimeoutMsec:               5000,
							Manual:                    true,
							DebugTimeBetweenStepsMsec: 1000,
						},
					},
					Utility: &configs.UtilityConfig{
						MaxMempoolTransactionBytes: 1073741824,
						MaxMempoolTransactions:     9000,
					},
					Persistence: &configs.PersistenceConfig{
						PostgresUrl:       "postgres://postgres:postgres@pocket-db:5432/postgres",
						NodeSchema:        "node1",
						BlockStorePath:    "/var/blockstore",
						TxIndexerPath:     "",
						TreesStoreDir:     "/var/trees",
						MaxConnsCount:     8,
						MinConnsCount:     0,
						MaxConnLifetime:   "1h",
						MaxConnIdleTime:   "30m",
						HealthCheckPeriod: "5m",
					},
					P2P: &configs.P2PConfig{
						PrivateKey:            "7e1cd00241f5a743dc1a42919ef56b0d32d8e868eae216c03d44c52faa415ef2f48654c9bffccd7a858dc5577551ff650f8df9f1ec5bb668f339f594f2380ba1",
						ConsensusPort:         8080,
						UseRainTree:           true,
						IsEmptyConnectionType: false,
						MaxMempoolCount:       1e5,
					},
					Telemetry: &configs.TelemetryConfig{
						Enabled:  true,
						Address:  "0.0.0.0:9000",
						Endpoint: "/metrics",
					},
					Logger: &configs.LoggerConfig{
						Level:  "debug",
						Format: "pretty",
					},
					RPC: &configs.RPCConfig{
						Enabled: true,
						Port:    "50832",
						Timeout: 30000,
						UseCors: false,
					},
				},
				genesisState: expectedGenesis,
				clock:        clock.New(),
			},
			assertion: func(tt require.TestingT, want, got interface{}, _ ...interface{}) {
				require.Equal(tt, want.(*Manager).config, got.(*Manager).config)
				require.Equal(tt, want.(*Manager).genesisState, got.(*Manager).genesisState)
			},
		},
		{
			name: "unset MaxMempoolCount should fallback to default value",
			args: args{
				configReader: strings.NewReader(string(`{
					"p2p": {
					  "consensus_port": 8080,
					  "use_rain_tree": true,
					  "is_empty_connection_type": false,
					  "private_key": "d936becdefdc21490dd1d1789c1887676ef2b17139526a68c79cfa66075fad2df511f0037512e802a584a1ef714790013f3db8d79e5f62cc2cae6902e1d7410b"
					}
				  }`)),
				genesisReader: strings.NewReader(string(buildGenesisBytes)),
			},
			want: &Manager{
				config: &configs.Config{
					P2P: &configs.P2PConfig{
						PrivateKey:            "d936becdefdc21490dd1d1789c1887676ef2b17139526a68c79cfa66075fad2df511f0037512e802a584a1ef714790013f3db8d79e5f62cc2cae6902e1d7410b",
						ConsensusPort:         8080,
						UseRainTree:           true,
						IsEmptyConnectionType: false,
						MaxMempoolCount:       defaults.DefaultP2PMaxMempoolCount,
					},
				},
				genesisState: expectedGenesis,
				clock:        clock.New(),
			},
			assertion: func(tt require.TestingT, want, got interface{}, _ ...interface{}) {
				require.Equal(tt, want.(*Manager).config.P2P, got.(*Manager).config.P2P)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewManagerFromReaders(tt.args.configReader, tt.args.genesisReader, tt.args.options...)
			tt.assertion(t, tt.want, got)
		})
	}
}
