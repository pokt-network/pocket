package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_       current_height_provider.CurrentHeightProvider = &rpcCurrentHeightProvider{}
	rpcHost string
)

func init() {
	// by default, we point at the same endpoint used by the CLI but the debug client is used either in docker-compose of K8S, therefore we cater for overriding
	rpcHost = runtime.GetEnv("RPC_HOST", defaults.Validator1EndpointK8S)
}

type rpcCurrentHeightProvider struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	rpcUrl    string
	rpcClient *rpc.ClientWithResponses
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(rpcCurrentHeightProvider).Create(bus, options...)
}

// Create implements current_height_provider.CurrentHeightProvider
func (*rpcCurrentHeightProvider) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return NewRPCCurrentHeightProvider(options...), nil
}

// GetModuleName implements current_height_provider.CurrentHeightProvider
func (*rpcCurrentHeightProvider) GetModuleName() string {
	return current_height_provider.ModuleName
}

func (dchp *rpcCurrentHeightProvider) CurrentHeight() uint64 {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)

	response, err := dchp.rpcClient.GetV1ConsensusStateWithResponse(ctx)
	if err != nil {
		cancel()
		log.Fatalf("could not get consensus state from RPC: %v", err)
	}
	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		cancel()
		log.Fatalf("error retrieving consensus state from RPC. Unexpected status code: %d", statusCode)
	}
	cancel()
	return uint64(response.JSONDefault.Height)
}

func NewRPCCurrentHeightProvider(options ...modules.ModuleOption) *rpcCurrentHeightProvider {
	dabp := &rpcCurrentHeightProvider{
		rpcUrl: fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort),
	}

	for _, o := range options {
		o(dabp)
	}

	initRPCClient(dabp)

	return dabp
}

func initRPCClient(dabp *rpcCurrentHeightProvider) {
	rpcClient, err := rpc.NewClientWithResponses(dabp.rpcUrl)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	dabp.rpcClient = rpcClient
}

// options

// WithCustomRPCUrl allows to specify a custom RPC URL
func WithCustomRPCUrl(rpcUrl string) modules.ModuleOption {
	return func(rabp modules.InitializableModule) {
		rabp.(*rpcCurrentHeightProvider).rpcUrl = rpcUrl
	}
}
