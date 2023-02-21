package test_artifacts

// Cross module imports are okay because this is only used for testing and not business logic
import (
	"fmt"
	"strconv"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygenerator"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IMPROVE: Generate a proper genesis suite in the future.
func NewGenesisState(numValidators, numServicers, numApplications, numFisherman int) (genesisState *genesis.GenesisState, validatorPrivateKeys []string) {
	apps, appsPrivateKeys := NewActors(coreTypes.ActorType_ACTOR_TYPE_APP, numApplications)
	vals, validatorPrivateKeys := NewActors(coreTypes.ActorType_ACTOR_TYPE_VAL, numValidators)
	servicers, snPrivateKeys := NewActors(coreTypes.ActorType_ACTOR_TYPE_SERVICER, numServicers)
	fish, fishPrivateKeys := NewActors(coreTypes.ActorType_ACTOR_TYPE_FISH, numFisherman)

	genesisState = &genesis.GenesisState{
		GenesisTime:   timestamppb.Now(),
		ChainId:       DefaultChainID,
		MaxBlockBytes: DefaultMaxBlockBytes,
		Pools:         NewPools(),
		Accounts:      NewAccounts(numValidators+numServicers+numApplications+numFisherman, append(append(append(validatorPrivateKeys, snPrivateKeys...), fishPrivateKeys...), appsPrivateKeys...)...), // TODO(olshansky): clean this up
		Applications:  apps,
		Validators:    vals,
		Servicers:     servicers,
		Fishermen:     fish,
		Params:        DefaultParams(),
	}

	// TODO: Generalize this to all actors and not just validators
	return genesisState, validatorPrivateKeys
}

func NewDefaultConfigs(privateKeys []string) (cfgs []*configs.Config) {
	for i, pk := range privateKeys {
		cfgs = append(cfgs, configs.NewDefaultConfig(
			configs.WithPK(pk),
			configs.WithNodeSchema("node"+strconv.Itoa(i+1)),
		))
	}
	return
}

// REFACTOR: Test artifact generator should reflect the sum of the initial account values to populate the initial pool values
func NewPools() (pools []*coreTypes.Account) {
	for _, name := range coreTypes.Pools_name {
		if name == coreTypes.Pools_POOLS_FEE_COLLECTOR.FriendlyName() {
			pools = append(pools, &coreTypes.Account{
				Address: name,
				Amount:  "0",
			})
			continue
		}
		pools = append(pools, &coreTypes.Account{
			Address: name,
			Amount:  DefaultAccountAmountString,
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
			Amount:  DefaultAccountAmountString,
		})
	}
	return
}

// TODO: The current implementation of NewActors  will have overlapping `ServiceUrl` for different
//
//	types of actors which needs to be fixed.
func NewActors(actorType coreTypes.ActorType, n int) (actors []*coreTypes.Actor, privateKeys []string) {
	for i := 0; i < n; i++ {
		genericParam := getServiceUrl(i + 1)
		if int32(actorType) == int32(coreTypes.ActorType_ACTOR_TYPE_APP) {
			genericParam = DefaultMaxRelaysString
		}
		actor, pk := NewDefaultActor(int32(actorType), genericParam)
		actors = append(actors, actor)
		privateKeys = append(privateKeys, pk)
	}

	return
}

func getServiceUrl(n int) string {
	return fmt.Sprintf(ServiceUrlFormat, n)
}

func NewDefaultActor(actorType int32, genericParam string) (actor *coreTypes.Actor, privateKey string) {
	privKey, pubKey, addr := keygenerator.GetInstance().Next()
	chains := DefaultChains
	if actorType == int32(coreTypes.ActorType_ACTOR_TYPE_VAL) {
		chains = nil
	} else if actorType == int32(coreTypes.ActorType_ACTOR_TYPE_APP) {
		genericParam = DefaultMaxRelaysString
	}
	return &coreTypes.Actor{
		Address:         addr,
		PublicKey:       pubKey,
		Chains:          chains,
		GenericParam:    genericParam,
		StakedAmount:    DefaultStakeAmountString,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          addr,
		ActorType:       coreTypes.ActorType(actorType),
	}, privKey
}
