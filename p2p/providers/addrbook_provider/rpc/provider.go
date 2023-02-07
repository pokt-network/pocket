package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ addrbook_provider.AddrBookProvider = &rpcAddrBookProvider{}

var rpcHost string = defaults.DefaultRemoteCLIURL // by default, we point at the same endpoint used by the CLI but the debug client is used either in docker-compose of K8S, therefore we cater for overriding

func init() {
	if os.Getenv("RPC_HOST") != "" {
		rpcHost = os.Getenv("RPC_HOST")
	}
}

type rpcAddrBookProvider struct {
	modules.BaseIntegratableModule
	modules.BaseInterruptableModule

	rpcUrl    string
	p2pCfg    *configs.P2PConfig
	rpcClient *rpc.ClientWithResponses

	connFactory typesP2P.ConnectionFactory
}

func NewRPCAddrBookProvider(options ...modules.ModuleOption) *rpcAddrBookProvider {
	dabp := &rpcAddrBookProvider{
		rpcUrl:      fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort),
		connFactory: transport.CreateDialer, // default connection factory, overridable with WithConnectionFactory()
	}

	for _, o := range options {
		o(dabp)
	}

	initRPCClient(dabp)

	return dabp
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(rpcAddrBookProvider).Create(bus, options...)
}

func (*rpcAddrBookProvider) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return NewRPCAddrBookProvider(options...), nil
}

func (*rpcAddrBookProvider) GetModuleName() string {
	return addrbook_provider.ModuleName
}

func (dabp *rpcAddrBookProvider) GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	var (
		h         int64              = int64(height)
		actorType rpc.ActorTypesEnum = "validator"
	)
	response, err := dabp.rpcClient.GetV1P2pAddressBookWithResponse(ctx, &rpc.GetV1P2pAddressBookParams{Height: &h, ActorType: &actorType})
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
			Address:      rpcActor.Address,
			PublicKey:    rpcActor.PublicKey,
			GenericParam: rpcActor.ServiceUrl,
			ActorType:    types.ActorType_ACTOR_TYPE_VAL,
		})
	}

	return addrbook_provider.ActorsToAddrBook(dabp, coreActors)
}

func (dabp *rpcAddrBookProvider) GetConnFactory() typesP2P.ConnectionFactory {
	return dabp.connFactory
}

func (dabp *rpcAddrBookProvider) GetP2PConfig() *configs.P2PConfig {
	if dabp.p2pCfg == nil {
		return dabp.GetBus().GetRuntimeMgr().GetConfig().P2P
	}
	return dabp.p2pCfg
}

func (dabp *rpcAddrBookProvider) SetConnectionFactory(connFactory typesP2P.ConnectionFactory) {
	dabp.connFactory = connFactory
}

func initRPCClient(dabp *rpcAddrBookProvider) {
	rpcClient, err := rpc.NewClientWithResponses(dabp.rpcUrl)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	dabp.rpcClient = rpcClient
}

// options

// WithP2PConfig allows to specify a custom P2P config
func WithP2PConfig(p2pCfg *configs.P2PConfig) modules.ModuleOption {
	return func(rabp modules.InitializableModule) {
		rabp.(*rpcAddrBookProvider).p2pCfg = p2pCfg
	}
}

// WithCustomRPCUrl allows to specify a custom RPC URL
func WithCustomRPCUrl(rpcUrl string) modules.ModuleOption {
	return func(rabp modules.InitializableModule) {
		rabp.(*rpcAddrBookProvider).rpcUrl = rpcUrl
	}
}
