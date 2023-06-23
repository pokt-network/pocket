package rpc

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ current_height_provider.CurrentHeightProvider = &rpcCurrentHeightProvider{}

type rpcCurrentHeightProvider struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	rpcURL    string
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
		rpcURL: flags.RemoteCLIURL,
	}

	for _, o := range options {
		o(rchp)
	}

	rchp.initRPCClient()

	return rchp
}

func (rchp *rpcCurrentHeightProvider) initRPCClient() {
	rpcClient, err := rpc.NewClientWithResponses(rchp.rpcURL)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	rchp.rpcClient = rpcClient
}

// options

// WithCustomRPCURL allows to specify a custom RPC URL
func WithCustomRPCURL(rpcURL string) modules.ModuleOption {
	return func(rabp modules.InjectableModule) {
		rabp.(*rpcCurrentHeightProvider).rpcURL = rpcURL
	}
}
