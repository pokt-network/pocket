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
	rpcHost = runtime.GetEnv("RPC_HOST", defaults.DefaultRPCHost)
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

func (rchp *rpcCurrentHeightProvider) CurrentHeight() uint64 {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)

	response, err := rchp.rpcClient.GetV1ConsensusStateWithResponse(ctx)
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
	rchp := &rpcCurrentHeightProvider{
		rpcUrl: fmt.Sprintf("http://%s:%s", rpcHost, defaults.DefaultRPCPort),
	}

	for _, o := range options {
		o(rchp)
	}

	rchp.initRPCClient()

	return rchp
}

func (rchp *rpcCurrentHeightProvider) initRPCClient() {
	rpcClient, err := rpc.NewClientWithResponses(rchp.rpcUrl)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	rchp.rpcClient = rpcClient
}

// options

// WithCustomRPCUrl allows to specify a custom RPC URL
func WithCustomRPCUrl(rpcUrl string) modules.ModuleOption {
	return func(rabp modules.InitializableModule) {
		rabp.(*rpcCurrentHeightProvider).rpcUrl = rpcUrl
	}
}
