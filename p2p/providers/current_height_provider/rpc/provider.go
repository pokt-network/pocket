package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ current_height_provider.CurrentHeightProvider = &rpcCurrentHeightProvider{}

var rpcHost string = defaults.DefaultRemoteCLIURL // by default, we point at the same endpoint used by the CLI but the debug client is used either in docker-compose of K8S, therefore we cater for overriding

func init() {
	if os.Getenv("RPC_HOST") != "" {
		rpcHost = os.Getenv("RPC_HOST")
	}
}

type rpcCurrentHeightProvider struct {
	modules.BaseIntegratableModule
	modules.BaseInterruptableModule

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
	defer cancel()

	response, err := dchp.rpcClient.GetV1ConsensusStateWithResponse(ctx)
	if err != nil {
		log.Fatalf("could not get consensus state from RPC: %v", err)
	}
	statusCode := response.StatusCode()
	if statusCode != http.StatusOK {
		log.Fatalf("error retrieving consensus state from RPC. Unexpected status code: %d", statusCode)
	}

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
