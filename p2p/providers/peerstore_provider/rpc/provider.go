package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_       peerstore_provider.PeerstoreProvider = &rpcPeerstoreProvider{}
	rpcHost string
)

func init() {
	// by default, we point at the same endpoint used by the CLI but the debug client is used either in docker-compose of K8S, therefore we cater for overriding
	rpcHost = runtime.GetEnv("RPC_HOST", defaults.DefaultRPCHost)
}

type rpcPeerstoreProvider struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	rpcURL    string
	p2pCfg    *configs.P2PConfig
	rpcClient *rpc.ClientWithResponses
}

func NewRPCPeerstoreProvider(options ...modules.ModuleOption) *rpcPeerstoreProvider {
	rabp := &rpcPeerstoreProvider{
		rpcURL: fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort), // TODO: Make port configurable
	}

	for _, o := range options {
		o(rabp)
	}

	rabp.initRPCClient()

	return rabp
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(rpcPeerstoreProvider).Create(bus, options...)
}

func (*rpcPeerstoreProvider) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return NewRPCPeerstoreProvider(options...), nil
}

func (*rpcPeerstoreProvider) GetModuleName() string {
	return peerstore_provider.ModuleName
}

func (rpcPSP *rpcPeerstoreProvider) GetStakedPeerstoreAtHeight(height uint64) (typesP2P.Peerstore, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	var (
		h         int64              = int64(height)
		actorType rpc.ActorTypesEnum = "validator"
	)
	response, err := rpcPSP.rpcClient.GetV1P2pStakedActorsAddressBookWithResponse(ctx, &rpc.GetV1P2pStakedActorsAddressBookParams{Height: &h, ActorType: &actorType})
	if err != nil {
		return nil, err
	}
	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("error retrieving addressbook via rpc. Unexpected status code: %d", statusCode)
	}

	rpcActors := response.JSON200.Actors
	var coreActors []*types.Actor
	for _, rpcActor := range rpcActors {
		coreActors = append(coreActors, &types.Actor{
			Address:    rpcActor.Address,
			PublicKey:  rpcActor.PublicKey,
			ServiceUrl: rpcActor.ServiceUrl,
			ActorType:  types.ActorType_ACTOR_TYPE_VAL,
		})
	}

	return peerstore_provider.ActorsToPeerstore(rpcPSP, coreActors)
}

func (rabp *rpcPeerstoreProvider) GetP2PConfig() *configs.P2PConfig {
	if rabp.p2pCfg == nil {
		return rabp.GetBus().GetRuntimeMgr().GetConfig().P2P
	}
	return rabp.p2pCfg
}

func (rpcPSP *rpcPeerstoreProvider) initRPCClient() {
	rpcClient, err := rpc.NewClientWithResponses(rpcPSP.rpcURL)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	rpcPSP.rpcClient = rpcClient
}

// options

// WithP2PConfig allows to specify a custom P2P config
func WithP2PConfig(p2pCfg *configs.P2PConfig) modules.ModuleOption {
	return func(rabp modules.InitializableModule) {
		rabp.(*rpcPeerstoreProvider).p2pCfg = p2pCfg
	}
}

// WithCustomRPCURL allows to specify a custom RPC URL
func WithCustomRPCURL(rpcURL string) modules.ModuleOption {
	return func(rabp modules.InitializableModule) {
		rabp.(*rpcPeerstoreProvider).rpcURL = rpcURL
	}
}
