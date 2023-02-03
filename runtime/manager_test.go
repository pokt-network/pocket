package runtime

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/runtime/configs"
	configTypes "github.com/pokt-network/pocket/runtime/configs/types"
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
			Address: "6f66574e1f50f0ef72dff748c3f11b9e0e89d32a",
			Amount:  "100000000000000",
		},
		{
			Address: "67eb3f0a50ae459fecf666be0e93176e92441317",
			Amount:  "100000000000000",
		},
		{
			Address: "3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99",
			Amount:  "100000000000000",
		},
		{
			Address: "113fdb095d42d6e09327ab5b8df13fd8197a1eaf",
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
			Address:         "113fdb095d42d6e09327ab5b8df13fd8197a1eaf",
			PublicKey:       "53ee26c82826694ffe1773d7b60d5f20dd9e91bdf8745544711bec5ff9c6fb4a",
			Chains:          nil,
			GenericParam:    "node1.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "113fdb095d42d6e09327ab5b8df13fd8197a1eaf",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99",
			PublicKey:       "a8b6be75d7551da093f788f7286c3a9cb885cfc8e52710eac5f1d5e5b4bf19b2",
			Chains:          nil,
			GenericParam:    "node2.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "67eb3f0a50ae459fecf666be0e93176e92441317",
			PublicKey:       "c16043323c83ffd901a8bf7d73543814b8655aa4695f7bfb49d01926fc161cdb",
			Chains:          nil,
			GenericParam:    "node3.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "67eb3f0a50ae459fecf666be0e93176e92441317",
		},
		{
			ActorType:       types.ActorType_ACTOR_TYPE_VAL,
			Address:         "6f66574e1f50f0ef72dff748c3f11b9e0e89d32a",
			PublicKey:       "b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4",
			Chains:          nil,
			GenericParam:    "node4.consensus:8080",
			StakedAmount:    "1000000000000",
			PausedHeight:    -1,
			UnstakingHeight: -1,
			Output:          "6f66574e1f50f0ef72dff748c3f11b9e0e89d32a",
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
					PrivateKey:    "c6c136d010d07d7f5e9944aa3594a10f9210dd3e26ebc1bc1516a6d957fd0df353ee26c82826694ffe1773d7b60d5f20dd9e91bdf8745544711bec5ff9c6fb4a",
					Consensus: &configs.ConsensusConfig{
						PrivateKey:      "c6c136d010d07d7f5e9944aa3594a10f9210dd3e26ebc1bc1516a6d957fd0df353ee26c82826694ffe1773d7b60d5f20dd9e91bdf8745544711bec5ff9c6fb4a",
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
						PrivateKey:      "c6c136d010d07d7f5e9944aa3594a10f9210dd3e26ebc1bc1516a6d957fd0df353ee26c82826694ffe1773d7b60d5f20dd9e91bdf8745544711bec5ff9c6fb4a",
						ConsensusPort:   8080,
						UseRainTree:     true,
						ConnectionType:  configTypes.ConnectionType_TCPConnection,
						MaxMempoolCount: 1e5,
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
			assertion: func(tt require.TestingT, want, got any, _ ...any) {
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
					  "private_key": "6fd0bc54cc2dd205eaf226eebdb0451629b321f11d279013ce6fdd5a33059256b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4"
					}
				  }`)),
				genesisReader: strings.NewReader(string(buildGenesisBytes)),
			},
			want: &Manager{
				config: &configs.Config{
					P2P: &configs.P2PConfig{
						PrivateKey:      "6fd0bc54cc2dd205eaf226eebdb0451629b321f11d279013ce6fdd5a33059256b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4",
						ConsensusPort:   8080,
						UseRainTree:     true,
						ConnectionType:  configTypes.ConnectionType_TCPConnection,
						MaxMempoolCount: defaults.DefaultP2PMaxMempoolCount,
					},
				},
				genesisState: expectedGenesis,
				clock:        clock.New(),
			},
			assertion: func(tt require.TestingT, want, got any, _ ...any) {
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
