package runtime

import (
	"io"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/converters"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
		name string
		args args
		want *Manager
	}{
		{
			name: "reading from the build directory",
			args: args{
				configReader:  strings.NewReader(string(buildConfigBytes)),
				genesisReader: strings.NewReader(string(buildGenesisBytes)),
			},
			want: &Manager{
				&configs.Config{
					RootDirectory: "/go/src/github.com/pocket-network",
					PrivateKey:    "6fd0bc54cc2dd205eaf226eebdb0451629b321f11d279013ce6fdd5a33059256b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4",
					Consensus: &configs.ConsensusConfig{
						PrivateKey:      "6fd0bc54cc2dd205eaf226eebdb0451629b321f11d279013ce6fdd5a33059256b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4",
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
						PostgresUrl:    "postgres://postgres:postgres@pocket-db:5432/postgres",
						NodeSchema:     "node1",
						BlockStorePath: "/var/blockstore",
						TxIndexerPath:  "",
						TreesStoreDir:  "/var/trees",
					},
					P2P: &configs.P2PConfig{
						PrivateKey:            "6fd0bc54cc2dd205eaf226eebdb0451629b321f11d279013ce6fdd5a33059256b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4",
						ConsensusPort:         8080,
						UseRainTree:           true,
						IsEmptyConnectionType: false,
						MaxMempoolCount:       1e6,
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
				&genesis.GenesisState{
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
							Address:         "6f66574e1f50f0ef72dff748c3f11b9e0e89d32a",
							PublicKey:       "b2eda2232ffb2750bf761141f70f75a03a025f65b2b2b417c7f8b3c9ca91e8e4",
							Chains:          nil,
							GenericParam:    "node1.consensus:8080",
							StakedAmount:    "1000000000000",
							PausedHeight:    -1,
							UnstakingHeight: -1,
							Output:          "6f66574e1f50f0ef72dff748c3f11b9e0e89d32a",
						},
						{
							ActorType:       types.ActorType_ACTOR_TYPE_VAL,
							Address:         "67eb3f0a50ae459fecf666be0e93176e92441317",
							PublicKey:       "c16043323c83ffd901a8bf7d73543814b8655aa4695f7bfb49d01926fc161cdb",
							Chains:          nil,
							GenericParam:    "node2.consensus:8080",
							StakedAmount:    "1000000000000",
							PausedHeight:    -1,
							UnstakingHeight: -1,
							Output:          "67eb3f0a50ae459fecf666be0e93176e92441317",
						},
						{
							ActorType:       types.ActorType_ACTOR_TYPE_VAL,
							Address:         "3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99",
							PublicKey:       "a8b6be75d7551da093f788f7286c3a9cb885cfc8e52710eac5f1d5e5b4bf19b2",
							Chains:          nil,
							GenericParam:    "node3.consensus:8080",
							StakedAmount:    "1000000000000",
							PausedHeight:    -1,
							UnstakingHeight: -1,
							Output:          "3f52e08c4b3b65ab7cf098d77df5bf8cedcf5f99",
						},
						{
							ActorType:       types.ActorType_ACTOR_TYPE_VAL,
							Address:         "113fdb095d42d6e09327ab5b8df13fd8197a1eaf",
							PublicKey:       "53ee26c82826694ffe1773d7b60d5f20dd9e91bdf8745544711bec5ff9c6fb4a",
							Chains:          nil,
							GenericParam:    "node4.consensus:8080",
							StakedAmount:    "1000000000000",
							PausedHeight:    -1,
							UnstakingHeight: -1,
							Output:          "113fdb095d42d6e09327ab5b8df13fd8197a1eaf",
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
					Params: &genesis.Params{
						BlocksPerSession:                         4,
						AppMinimumStake:                          converters.BigIntToString(big.NewInt(15000000000)),
						AppMaxChains:                             15,
						AppBaselineStakeRate:                     100,
						AppStakingAdjustment:                     0,
						AppUnstakingBlocks:                       2016,
						AppMinimumPauseBlocks:                    4,
						AppMaxPauseBlocks:                        672,
						ServiceNodeMinimumStake:                  converters.BigIntToString(big.NewInt(15000000000)),
						ServiceNodeMaxChains:                     15,
						ServiceNodeUnstakingBlocks:               2016,
						ServiceNodeMinimumPauseBlocks:            4,
						ServiceNodeMaxPauseBlocks:                672,
						ServiceNodesPerSession:                   24,
						FishermanMinimumStake:                    converters.BigIntToString(big.NewInt(15000000000)),
						FishermanMaxChains:                       15,
						FishermanUnstakingBlocks:                 2016,
						FishermanMinimumPauseBlocks:              4,
						FishermanMaxPauseBlocks:                  672,
						ValidatorMinimumStake:                    converters.BigIntToString(big.NewInt(15000000000)),
						ValidatorUnstakingBlocks:                 2016,
						ValidatorMinimumPauseBlocks:              4,
						ValidatorMaxPauseBlocks:                  672,
						ValidatorMaximumMissedBlocks:             5,
						ValidatorMaxEvidenceAgeInBlocks:          8,
						ProposerPercentageOfFees:                 10,
						MissedBlocksBurnPercentage:               1,
						DoubleSignBurnPercentage:                 5,
						MessageDoubleSignFee:                     converters.BigIntToString(big.NewInt(10000)),
						MessageSendFee:                           converters.BigIntToString(big.NewInt(10000)),
						MessageStakeFishermanFee:                 converters.BigIntToString(big.NewInt(10000)),
						MessageEditStakeFishermanFee:             converters.BigIntToString(big.NewInt(10000)),
						MessageUnstakeFishermanFee:               converters.BigIntToString(big.NewInt(10000)),
						MessagePauseFishermanFee:                 converters.BigIntToString(big.NewInt(10000)),
						MessageUnpauseFishermanFee:               converters.BigIntToString(big.NewInt(10000)),
						MessageFishermanPauseServiceNodeFee:      converters.BigIntToString(big.NewInt(10000)),
						MessageTestScoreFee:                      converters.BigIntToString(big.NewInt(10000)),
						MessageProveTestScoreFee:                 converters.BigIntToString(big.NewInt(10000)),
						MessageStakeAppFee:                       converters.BigIntToString(big.NewInt(10000)),
						MessageEditStakeAppFee:                   converters.BigIntToString(big.NewInt(10000)),
						MessageUnstakeAppFee:                     converters.BigIntToString(big.NewInt(10000)),
						MessagePauseAppFee:                       converters.BigIntToString(big.NewInt(10000)),
						MessageUnpauseAppFee:                     converters.BigIntToString(big.NewInt(10000)),
						MessageStakeValidatorFee:                 converters.BigIntToString(big.NewInt(10000)),
						MessageEditStakeValidatorFee:             converters.BigIntToString(big.NewInt(10000)),
						MessageUnstakeValidatorFee:               converters.BigIntToString(big.NewInt(10000)),
						MessagePauseValidatorFee:                 converters.BigIntToString(big.NewInt(10000)),
						MessageUnpauseValidatorFee:               converters.BigIntToString(big.NewInt(10000)),
						MessageStakeServiceNodeFee:               converters.BigIntToString(big.NewInt(10000)),
						MessageEditStakeServiceNodeFee:           converters.BigIntToString(big.NewInt(10000)),
						MessageUnstakeServiceNodeFee:             converters.BigIntToString(big.NewInt(10000)),
						MessagePauseServiceNodeFee:               converters.BigIntToString(big.NewInt(10000)),
						MessageUnpauseServiceNodeFee:             converters.BigIntToString(big.NewInt(10000)),
						MessageChangeParameterFee:                converters.BigIntToString(big.NewInt(10000)),
						AclOwner:                                 "da034209758b78eaea06dd99c07909ab54c99b45",
						BlocksPerSessionOwner:                    "da034209758b78eaea06dd99c07909ab54c99b45",
						AppMinimumStakeOwner:                     "da034209758b78eaea06dd99c07909ab54c99b45",
						AppMaxChainsOwner:                        "da034209758b78eaea06dd99c07909ab54c99b45",
						AppBaselineStakeRateOwner:                "da034209758b78eaea06dd99c07909ab54c99b45",
						AppStakingAdjustmentOwner:                "da034209758b78eaea06dd99c07909ab54c99b45",
						AppUnstakingBlocksOwner:                  "da034209758b78eaea06dd99c07909ab54c99b45",
						AppMinimumPauseBlocksOwner:               "da034209758b78eaea06dd99c07909ab54c99b45",
						AppMaxPausedBlocksOwner:                  "da034209758b78eaea06dd99c07909ab54c99b45",
						ServiceNodeMinimumStakeOwner:             "da034209758b78eaea06dd99c07909ab54c99b45",
						ServiceNodeMaxChainsOwner:                "da034209758b78eaea06dd99c07909ab54c99b45",
						ServiceNodeUnstakingBlocksOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						ServiceNodeMinimumPauseBlocksOwner:       "da034209758b78eaea06dd99c07909ab54c99b45",
						ServiceNodeMaxPausedBlocksOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						ServiceNodesPerSessionOwner:              "da034209758b78eaea06dd99c07909ab54c99b45",
						FishermanMinimumStakeOwner:               "da034209758b78eaea06dd99c07909ab54c99b45",
						FishermanMaxChainsOwner:                  "da034209758b78eaea06dd99c07909ab54c99b45",
						FishermanUnstakingBlocksOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						FishermanMinimumPauseBlocksOwner:         "da034209758b78eaea06dd99c07909ab54c99b45",
						FishermanMaxPausedBlocksOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						ValidatorMinimumStakeOwner:               "da034209758b78eaea06dd99c07909ab54c99b45",
						ValidatorUnstakingBlocksOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						ValidatorMinimumPauseBlocksOwner:         "da034209758b78eaea06dd99c07909ab54c99b45",
						ValidatorMaxPausedBlocksOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						ValidatorMaximumMissedBlocksOwner:        "da034209758b78eaea06dd99c07909ab54c99b45",
						ValidatorMaxEvidenceAgeInBlocksOwner:     "da034209758b78eaea06dd99c07909ab54c99b45",
						ProposerPercentageOfFeesOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MissedBlocksBurnPercentageOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						DoubleSignBurnPercentageOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageDoubleSignFeeOwner:                "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageSendFeeOwner:                      "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageStakeFishermanFeeOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageEditStakeFishermanFeeOwner:        "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnstakeFishermanFeeOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						MessagePauseFishermanFeeOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnpauseFishermanFeeOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageFishermanPauseServiceNodeFeeOwner: "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageTestScoreFeeOwner:                 "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageProveTestScoreFeeOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageStakeAppFeeOwner:                  "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageEditStakeAppFeeOwner:              "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnstakeAppFeeOwner:                "da034209758b78eaea06dd99c07909ab54c99b45",
						MessagePauseAppFeeOwner:                  "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnpauseAppFeeOwner:                "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageStakeValidatorFeeOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageEditStakeValidatorFeeOwner:        "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnstakeValidatorFeeOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						MessagePauseValidatorFeeOwner:            "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnpauseValidatorFeeOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageStakeServiceNodeFeeOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageEditStakeServiceNodeFeeOwner:      "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnstakeServiceNodeFeeOwner:        "da034209758b78eaea06dd99c07909ab54c99b45",
						MessagePauseServiceNodeFeeOwner:          "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageUnpauseServiceNodeFeeOwner:        "da034209758b78eaea06dd99c07909ab54c99b45",
						MessageChangeParameterFeeOwner:           "da034209758b78eaea06dd99c07909ab54c99b45",
					},
				},
				clock.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewManagerFromReaders(tt.args.configReader, tt.args.genesisReader, tt.args.options...)
			require.Equal(t, tt.want, got)
		})
	}
}
