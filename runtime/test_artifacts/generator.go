package test_artifacts

// Cross module imports are okay because this is only used for testing and not business logic
import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	"github.com/pokt-network/pocket/shared/core/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GenesisOption func(*genesis.GenesisState)

// IMPROVE: Generate a proper genesis suite in the future.
func NewGenesisState(numValidators, numServicers, numApplications, numFisherman int, genesisOpts ...GenesisOption) (genesisState *genesis.GenesisState, validatorPrivateKeys []string) {
	applications, appPrivateKeys := newActors(coreTypes.ActorType_ACTOR_TYPE_APP, numApplications)
	validators, validatorPrivateKeys := newActors(coreTypes.ActorType_ACTOR_TYPE_VAL, numValidators)
	servicers, servicerPrivateKeys := newActors(coreTypes.ActorType_ACTOR_TYPE_SERVICER, numServicers)
	fishermen, fishPrivateKeys := newActors(coreTypes.ActorType_ACTOR_TYPE_FISH, numFisherman)

	allActorsKeys := append(append(append(validatorPrivateKeys, servicerPrivateKeys...), fishPrivateKeys...), appPrivateKeys...)
	allActorAccounts := newAccountsWithKeys(allActorsKeys)

	genesisState = &genesis.GenesisState{
		GenesisTime:   timestamppb.Now(),
		ChainId:       DefaultChainID,
		MaxBlockBytes: DefaultMaxBlockBytes,
		Pools:         NewPools(),
		Accounts:      allActorAccounts,
		Applications:  applications,
		Validators:    validators,
		Servicers:     servicers,
		Fishermen:     fishermen,
		Params:        DefaultParams(),
	}

	for _, o := range genesisOpts {
		o(genesisState)
	}

	// TECHDEBT: Generalize this to all actors and not just validators
	return genesisState, validatorPrivateKeys
}

func WithActors(actors []*coreTypes.Actor, actorKeys []string) func(*genesis.GenesisState) {
	return func(genesis *genesis.GenesisState) {
		newActorAccounts := newAccountsWithKeys(actorKeys)
		genesis.Accounts = append(genesis.Accounts, newActorAccounts...)
		for _, actor := range actors {
			switch actor.ActorType {
			case types.ActorType_ACTOR_TYPE_APP:
				genesis.Applications = append(genesis.Applications, actor)
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				genesis.Validators = append(genesis.Validators, actor)
			case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
				genesis.Servicers = append(genesis.Servicers, actor)
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				genesis.Fishermen = append(genesis.Fishermen, actor)
			default:
				panic(fmt.Sprintf("invalid actor type: %s", actor.ActorType))
			}
		}
	}
}

func NewDefaultConfigs(privateKeys []string) (cfgs []*configs.Config) {
	for i, pk := range privateKeys {
		postgresSchema := "node" + strconv.Itoa(i+1)
		cfgs = append(cfgs, configs.NewDefaultConfig(
			configs.WithPK(pk),
			configs.WithNodeSchema(postgresSchema),
		))
	}
	return cfgs
}

func NewPools() (pools []*coreTypes.Account) {
	for _, value := range coreTypes.Pools_value {
		if value == int32(coreTypes.Pools_POOLS_UNSPECIFIED) {
			continue
		}

		// TECHDEBT: Test artifact should reflect the sum of the initial account values
		// rather than be set to `DefaultAccountAmountString`
		amount := DefaultAccountAmountString
		if value == int32(coreTypes.Pools_POOLS_FEE_COLLECTOR) {
			amount = "0" // fees are empty at genesis
		}

		poolAddr := hex.EncodeToString(coreTypes.Pools(value).Address())

		pools = append(pools, &coreTypes.Account{
			Address: poolAddr,
			Amount:  amount,
		})
	}
	return pools
}

func newAccountsWithKeys(privateKeys []string) (accounts []*coreTypes.Account) {
	for _, pk := range privateKeys {
		pk, _ := crypto.NewPrivateKey(pk)
		addr := pk.Address().String()
		accounts = append(accounts, &coreTypes.Account{
			Address: addr,
			Amount:  DefaultAccountAmountString,
		})
	}
	return accounts
}

func newAccounts(numActors int) (accounts []*coreTypes.Account) {
	for i := 0; i < numActors; i++ {
		_, _, addr := keygen.GetInstance().Next()
		accounts = append(accounts, &coreTypes.Account{
			Address: addr,
			Amount:  DefaultAccountAmountString,
		})
	}
	return accounts
}

// TECHDEBT: Current implementation of `newActors` will result in non-unique ServiceURLs if called
// more than once.
func newActors(actorType coreTypes.ActorType, numActors int) (actors []*coreTypes.Actor, privateKeys []string) {
	chains := DefaultChains
	if actorType == coreTypes.ActorType_ACTOR_TYPE_VAL {
		chains = nil
	}
	for i := 0; i < numActors; i++ {
		serviceURL := getServiceURL(i + 1)
		actor, pk := NewDefaultActor(actorType, serviceURL, chains)
		actors = append(actors, actor)
		privateKeys = append(privateKeys, pk)
	}
	return actors, privateKeys
}

func NewDefaultActor(actorType coreTypes.ActorType, serviceURL string, chains []string) (actor *coreTypes.Actor, privateKey string) {
	privKey, pubKey, addr := keygen.GetInstance().Next()
	return &coreTypes.Actor{
		ActorType:       actorType,
		Address:         addr,
		PublicKey:       pubKey,
		Chains:          chains,
		ServiceUrl:      serviceURL,
		StakedAmount:    DefaultStakeAmountString,
		PausedHeight:    DefaultPauseHeight,
		UnstakingHeight: DefaultUnstakingHeight,
		Output:          addr,
	}, privKey
}

func getServiceURL(n int) string {
	return fmt.Sprintf(ServiceURLFormat, n)
}
