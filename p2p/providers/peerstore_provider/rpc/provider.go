package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pokt-network/pocket/app/client/cli/flags"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/rpc"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ peerstore_provider.PeerstoreProvider = &rpcPeerstoreProvider{}
	_ rpcPeerstoreProviderFactory          = &rpcPeerstoreProvider{}
)

type rpcPeerstoreProviderOption func(*rpcPeerstoreProvider)
type rpcPeerstoreProviderFactory = modules.FactoryWithOptions[peerstore_provider.PeerstoreProvider, rpcPeerstoreProviderOption]

type rpcPeerstoreProvider struct {
	base_modules.IntegrableModule

	rpcURL    string
	rpcClient *rpc.ClientWithResponses
}

func Create(
	bus modules.Bus,
	options ...rpcPeerstoreProviderOption,
) (peerstore_provider.PeerstoreProvider, error) {
	return new(rpcPeerstoreProvider).Create(bus, options...)
}

func (*rpcPeerstoreProvider) Create(
	bus modules.Bus,
	options ...rpcPeerstoreProviderOption,
) (peerstore_provider.PeerstoreProvider, error) {
	rpcPSP := &rpcPeerstoreProvider{
		rpcURL: flags.RemoteCLIURL,
	}
	bus.RegisterModule(rpcPSP)

	for _, o := range options {
		o(rpcPSP)
	}

	rpcPSP.initRPCClient()

	return rpcPSP, nil
}

func (*rpcPeerstoreProvider) GetModuleName() string {
	return peerstore_provider.PeerstoreProviderSubmoduleName
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

func (rpcPSP *rpcPeerstoreProvider) GetUnstakedPeerstore() (typesP2P.Peerstore, error) {
	// TECHDEBT(#810, #811): use `bus.GetUnstakedActorRouter()` once it's available.
	unstakedActorRouterMod, err := rpcPSP.GetBus().GetModulesRegistry().GetModule(typesP2P.UnstakedActorRouterSubmoduleName)
	if err != nil {
		return nil, err
	}

	unstakedActorRouter, ok := unstakedActorRouterMod.(typesP2P.Router)
	if !ok {
		return nil, fmt.Errorf("unexpected unstaked actor router submodule type: %T", unstakedActorRouterMod)
	}

	return unstakedActorRouter.GetPeerstore(), nil
}

func (rpcPSP *rpcPeerstoreProvider) initRPCClient() {
	rpcClient, err := rpc.NewClientWithResponses(rpcPSP.rpcURL)
	if err != nil {
		log.Fatalf("could not create RPC client: %v", err)
	}
	rpcPSP.rpcClient = rpcClient
}

// options

// WithCustomRPCURL allows to specify a custom RPC URL
func WithCustomRPCURL(rpcURL string) rpcPeerstoreProviderOption {
	return func(rpcPSP *rpcPeerstoreProvider) {
		rpcPSP.rpcURL = rpcURL
	}
}
