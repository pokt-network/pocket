package test_artifacts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	DefaultChains              = []string{"0001"}
	DefaultServiceURL          = ""
	DefaultStakeAmount         = big.NewInt(1000000000000)
	DefaultStakeAmountString   = types.BigIntToString(DefaultStakeAmount)
	DefaultMaxRelays           = big.NewInt(1000000)
	DefaultMaxRelaysString     = types.BigIntToString(DefaultMaxRelays)
	DefaultAccountAmount       = big.NewInt(100000000000000)
	DefaultAccountAmountString = types.BigIntToString(DefaultAccountAmount)
	DefaultPauseHeight         = int64(-1)
	DefaultUnstakingHeight     = int64(-1)
	DefaultChainID             = "testnet"
	DefaultMaxBlockBytes       = uint64(4000000)
	DefaultRpcPort             = "26657"
	DefaultRpcTimeout          = uint64(3000)
	DefaultRemoteCliUrl        = "http://localhost:26657"
)

// TODO(drewsky): this is meant to be a **temporary** replacement for the recently deprecated
//
//	'genesis config' option. We need to implement a real suite soon!
func NewGenesisState(numValidators, numServiceNodes, numApplications, numFisherman int) (genesisState *genesis.GenesisState, validatorPrivateKeys []string) {
	apps, appsPrivateKeys := NewActors(genesis.ActorType_App, numApplications)
	vals, validatorPrivateKeys := NewActors(genesis.ActorType_Val, numValidators)
	serviceNodes, snPrivateKeys := NewActors(genesis.ActorType_Node, numServiceNodes)
	fish, fishPrivateKeys := NewActors(genesis.ActorType_Fish, numFisherman)
	return &genesis.GenesisState{
		Consensus: &genesis.ConsensusGenesisState{
			GenesisTime:   timestamppb.Now(),
			ChainId:       DefaultChainID,
			MaxBlockBytes: DefaultMaxBlockBytes,
		},
		Utility: &genesis.UtilityGenesisState{
			Pools:        NewPools(),
			Accounts:     NewAccounts(numValidators+numServiceNodes+numApplications+numFisherman, append(append(append(validatorPrivateKeys, snPrivateKeys...), fishPrivateKeys...), appsPrivateKeys...)...), // TODO(olshansky): clean this up
			Applications: apps,
			Validators:   vals,
			ServiceNodes: serviceNodes,
			Fishermen:    fish,
			Params:       DefaultParams(),
		},
	}, validatorPrivateKeys
}

func NewDefaultConfigs(privateKeys []string) (configs []*genesis.Config) {
	for i, pk := range privateKeys {
		configs = append(configs, NewDefaultConfig(i, pk))
	}
	return
}

func NewDefaultConfig(nodeNum int, privateKey string) *genesis.Config {
	return &genesis.Config{
		Base: &genesis.BaseConfig{
			RootDirectory: "/go/src/github.com/pocket-network",
			PrivateKey:    privateKey,
		},
		Consensus: &genesis.ConsensusConfig{
			MaxMempoolBytes: 500000000,
			PacemakerConfig: &genesis.PacemakerConfig{
				TimeoutMsec:               5000,
				Manual:                    true,
				DebugTimeBetweenStepsMsec: 1000,
			},
		},
		Utility: &genesis.UtilityConfig{},
		Persistence: &genesis.PersistenceConfig{
			PostgresUrl:    "postgres://postgres:postgres@pocket-db:5432/postgres",
			NodeSchema:     "node" + strconv.Itoa(nodeNum+1),
			BlockStorePath: "/var/blockstore",
		},
		P2P: &genesis.P2PConfig{
			ConsensusPort:  8080,
			UseRainTree:    true,
			ConnectionType: genesis.ConnectionType_TCPConnection,
		},
		Telemetry: &genesis.TelemetryConfig{
			Enabled:  true,
			Address:  "0.0.0.0:9000",
			Endpoint: "/metrics",
		},
		Rpc: &genesis.RPCConfig{
			Enabled:      true,
			Port:         DefaultRpcPort,
			Timeout:      DefaultRpcTimeout,
			RemoteCliUrl: DefaultRemoteCliUrl,
		},
	}
}

// TODO: in the real testing suite, we need to populate the pool amounts dependent on the actors
func NewPools() (pools []*genesis.Account) {
	for _, name := range genesis.Pool_Names_name {
		if name == genesis.Pool_Names_FeeCollector.String() {
			pools = append(pools, &genesis.Account{
				Address: name,
				Amount:  "0",
			})
			continue
		}
		pools = append(pools, &genesis.Account{
			Address: name,
			Amount:  DefaultAccountAmountString,
		})
	}
	return
}

func NewAccounts(n int, privateKeys ...string) (accounts []*genesis.Account) {
	for i := 0; i < n; i++ {
		_, _, addr := GenerateNewKeysStrings()
		if privateKeys != nil {
			pk, _ := crypto.NewPrivateKey(privateKeys[i])
			addr = pk.Address().String()
		}
		accounts = append(accounts, &genesis.Account{
			Address: addr,
			Amount:  DefaultAccountAmountString,
		})
	}
	return
}

func NewActors(actorType genesis.ActorType, n int) (actors []*genesis.Actor, privateKeys []string) {
	for i := 0; i < n; i++ {
		genericParam := fmt.Sprintf("node%d.consensus:8080", i+1)
		if actorType == genesis.ActorType_App {
			genericParam = DefaultMaxRelaysString
		}
		actor, pk := NewDefaultActor(actorType, genericParam)
		actors = append(actors, actor)
		privateKeys = append(privateKeys, pk)
	}
	return
}

func NewDefaultActor(actorType genesis.ActorType, genericParam string) (actor *genesis.Actor, privateKey string) {
	privKey, pubKey, addr := GenerateNewKeysStrings()
	chains := DefaultChains
	if actorType == genesis.ActorType_Val {
		chains = nil
	} else if actorType == genesis.ActorType_App {
		genericParam = DefaultMaxRelaysString
	}
	return &genesis.Actor{
		Address:         addr,
		PublicKey:       pubKey,
		Chains:          chains,
		GenericParam:    genericParam,
		StakedAmount:    DefaultStakeAmountString,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          addr,
		ActorType:       actorType,
	}, privKey
}

func GenerateNewKeys() (privateKey crypto.PrivateKey, publicKey crypto.PublicKey, address crypto.Address) {
	privateKey, _ = crypto.GeneratePrivateKey()
	publicKey = privateKey.PublicKey()
	address = publicKey.Address()
	return
}

func GenerateNewKeysStrings() (privateKey, publicKey, address string) {
	privKey, pubKey, addr := GenerateNewKeys()
	privateKey = privKey.String()
	publicKey = pubKey.String()
	address = addr.String()
	return
}

func ReadConfigAndGenesisFiles(configPath string, genesisPath string) (config *genesis.Config, genesisState *genesis.GenesisState) {
	if configPath == "" {
		log.Fatalf("config path cannot be empty")
	}
	if genesisPath == "" {
		log.Fatalf("genesis path cannot be empty")
	}

	config = new(genesis.Config)
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("[ERROR] an error occurred reading config.json file: %v", err.Error())
	}
	if err = json.Unmarshal(configFile, config); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the config.json file: %v", err.Error())
	}

	genesisState = new(genesis.GenesisState)
	genesisFile, err := ioutil.ReadFile(genesisPath)
	if err != nil {
		log.Fatalf("[ERROR] an error occurred reading genesis.json file: %v", err.Error())
	}
	if err = json.Unmarshal(genesisFile, genesisState); err != nil {
		log.Fatalf("[ERROR] an error occurred unmarshalling the genesis.json file: %v", err.Error())
	}
	return
}
