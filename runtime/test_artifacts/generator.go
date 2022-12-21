package test_artifacts

// Cross module imports are okay because this is only used for testing and not business logic
import (
	"fmt"
	"strconv"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygenerator"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IMPROVE: Generate a proper genesis suite in the future.
func NewGenesisState(numValidators, numServiceNodes, numApplications, numFisherman int) (*genesis.GenesisState, []string) {
	apps, appsPrivateKeys := NewActors(types.ActorType_App, numApplications)
	vals, validatorPrivateKeys := NewActors(types.ActorType_Validator, numValidators)
	serviceNodes, snPrivateKeys := NewActors(types.ActorType_ServiceNode, numServiceNodes)
	fish, fishPrivateKeys := NewActors(types.ActorType_Fisherman, numFisherman)

	genesisState := &genesis.GenesisState{
		GenesisTime:   timestamppb.Now(),
		ChainId:       defaults.DefaultChainID,
		MaxBlockBytes: defaults.DefaultMaxBlockBytes,
		Pools:         NewPools(),
		Accounts:      NewAccounts(numValidators+numServiceNodes+numApplications+numFisherman, append(append(append(validatorPrivateKeys, snPrivateKeys...), fishPrivateKeys...), appsPrivateKeys...)...), // TODO(olshansky): clean this up
		Applications:  apps,
		Validators:    vals,
		ServiceNodes:  serviceNodes,
		Fishermen:     fish,
		Params:        DefaultParams(),
	}

	// TODO: Generalize this to all actors and not just validators
	return genesisState, validatorPrivateKeys
}

func NewDefaultConfigs(privateKeys []string) (configs []*configs.Config) {
	for i, pk := range privateKeys {
		configs = append(configs, NewDefaultConfig(i, pk))
	}
	return
}

func NewDefaultConfig(i int, pk string) *configs.Config {
	return &configs.Config{
		RootDirectory: "/go/src/github.com/pocket-network",
		PrivateKey:    pk,
		Consensus: &configs.ConsensusConfig{
			MaxMempoolBytes: 500000000,
			PacemakerConfig: &configs.PacemakerConfig{
				TimeoutMsec:               5000,
				Manual:                    true,
				DebugTimeBetweenStepsMsec: 1000,
			},
			PrivateKey: pk,
		},
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1024 * 1024 * 1024, // 1GB V0 defaults
			MaxMempoolTransactions:     9000,
		},
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:    "postgres://postgres:postgres@pocket-db:5432/postgres",
			NodeSchema:     "node" + strconv.Itoa(i+1),
			BlockStorePath: "/var/blockstore",
		},
		P2P: &configs.P2PConfig{
			ConsensusPort:         8080,
			UseRainTree:           true,
			IsEmptyConnectionType: false,
			PrivateKey:            pk,
		},
		Telemetry: &configs.TelemetryConfig{
			Enabled:  true,
			Address:  "0.0.0.0:9000",
			Endpoint: "/metrics",
		},
	}
}

func NewPools() (pools []*coreTypes.Account) { // TODO (Team) in the real testing suite, we need to populate the pool amounts dependent on the actors
	for _, name := range coreTypes.PoolNames_name {
		if name == coreTypes.PoolNames_POOL_NAMES_FEE_COLLECTOR.String() {
			pools = append(pools, &coreTypes.Account{
				Address: name,
				Amount:  "0",
			})
			continue
		}
		pools = append(pools, &coreTypes.Account{
			Address: name,
			Amount:  defaults.DefaultAccountAmountString,
		})
	}
	return
}

func NewAccounts(n int, privateKeys ...string) (accounts []*coreTypes.Account) {
	for i := 0; i < n; i++ {
		_, _, addr := keygenerator.GetInstance().Next()
		if privateKeys != nil {
			pk, _ := crypto.NewPrivateKey(privateKeys[i])
			addr = pk.Address().String()
		}
		accounts = append(accounts, &coreTypes.Account{
			Address: addr,
			Amount:  defaults.DefaultAccountAmountString,
		})
	}
	return
}

// TODO: The current implementation of NewActors  will have overlapping `ServiceUrl` for different
//       types of actors which needs to be fixed.
func NewActors(actorType typesUtil.ActorType, n int) (actors []*coreTypes.Actor, privateKeys []string) {
	for i := 0; i < n; i++ {
		genericParam := getServiceUrl(i + 1)
		if int32(actorType) == int32(types.ActorType_App) {
			genericParam = defaults.DefaultMaxRelaysString
		}
		actor, pk := NewDefaultActor(int32(actorType), genericParam)
		actors = append(actors, actor)
		privateKeys = append(privateKeys, pk)
	}

	return
}

func getServiceUrl(n int) string {
	return fmt.Sprintf(defaults.ServiceUrlFormat, n)
}

func NewDefaultActor(actorType int32, genericParam string) (actor *coreTypes.Actor, privateKey string) {
	privKey, pubKey, addr := keygenerator.GetInstance().Next()
	chains := defaults.DefaultChains
	if actorType == int32(coreTypes.ActorType_ACTOR_TYPE_VAL) {
		chains = nil
	} else if actorType == int32(types.ActorType_App) {
		genericParam = defaults.DefaultMaxRelaysString
	}
	return &coreTypes.Actor{
		Address:         addr,
		PublicKey:       pubKey,
		Chains:          chains,
		GenericParam:    genericParam,
		StakedAmount:    defaults.DefaultStakeAmountString,
		PausedHeight:    defaults.DefaultPauseHeight,
		UnstakingHeight: defaults.DefaultUnstakingHeight,
		Output:          addr,
		ActorType:       coreTypes.ActorType(actorType),
	}, privKey
}
