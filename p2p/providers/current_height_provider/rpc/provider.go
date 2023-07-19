package rpc

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ modules.CurrentHeightProvider = &rpcCurrentHeightProvider{}

type rpcCurrentHeightProvider struct {
	base_modules.IntegrableModule

	rpcURL    string
	rpcClient *rpc.ClientWithResponses
}

func Create(
	bus modules.Bus,
	options ...modules.CurrentHeightProviderOption,
) (modules.CurrentHeightProvider, error) {
	return new(rpcCurrentHeightProvider).Create(bus, options...)
}

// Create implements current_height_provider.CurrentHeightProvider
func (*rpcCurrentHeightProvider) Create(
	bus modules.Bus,
	options ...modules.CurrentHeightProviderOption,
) (modules.CurrentHeightProvider, error) {
	rpcHeightProvider := &rpcCurrentHeightProvider{
		rpcURL: flags.RemoteCLIURL,
	}
	bus.RegisterModule(rpcHeightProvider)

	for _, o := range options {
		o(rpcHeightProvider)
	}

	rpcHeightProvider.initRPCClient()

	return rpcHeightProvider, nil
}

// GetModuleName implements current_height_provider.CurrentHeightProvider
func (*rpcCurrentHeightProvider) GetModuleName() string {
	return modules.CurrentHeightProviderSubmoduleName
}

func (rpcCHP *rpcCurrentHeightProvider) CurrentHeight() uint64 {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)

	response, err := rpcCHP.rpcClient.GetV1ConsensusStateWithResponse(ctx)
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

func (rpcCHP *rpcCurrentHeightProvider) initRPCClient() {
	rpcClient, err := rpc.NewClientWithResponses(rpcCHP.rpcURL)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	rpcCHP.rpcClient = rpcClient
}

// options

// WithCustomRPCURL allows to specify a custom RPC URL
func WithCustomRPCURL(rpcURL string) modules.CurrentHeightProviderOption {
	return func(chp modules.CurrentHeightProvider) {
		chp.(*rpcCurrentHeightProvider).rpcURL = rpcURL
	}
}
