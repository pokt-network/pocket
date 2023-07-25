package persistence

import (
	"fmt"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ peerstore_provider.PeerstoreProvider = &persistencePeerstoreProvider{}
	_ persistencePStoreProviderFactory     = &persistencePeerstoreProvider{}
)

type (
	persistencePStoreProviderOption  func(*persistencePeerstoreProvider)
	persistencePStoreProviderFactory = modules.FactoryWithOptions[peerstore_provider.PeerstoreProvider, persistencePStoreProviderOption]
)

type persistencePeerstoreProvider struct {
	base_modules.IntegrableModule
}

func Create(bus modules.Bus, options ...persistencePStoreProviderOption) (peerstore_provider.PeerstoreProvider, error) {
	return new(persistencePeerstoreProvider).Create(bus, options...)
}

func (*persistencePeerstoreProvider) Create(bus modules.Bus, options ...persistencePStoreProviderOption) (peerstore_provider.PeerstoreProvider, error) {
	persistencePSP := &persistencePeerstoreProvider{
		IntegrableModule: *base_modules.NewIntegrableModule(bus),
	}
	bus.RegisterModule(persistencePSP)

	for _, o := range options {
		o(persistencePSP)
	}

	return persistencePSP, nil
}

func (*persistencePeerstoreProvider) GetModuleName() string {
	return peerstore_provider.PeerstoreProviderSubmoduleName
}

// GetStakedPeerstoreAtHeight implements the respective `PeerstoreProvider` interface method.
func (persistencePSP *persistencePeerstoreProvider) GetStakedPeerstoreAtHeight(height uint64) (typesP2P.Peerstore, error) {
	readCtx, err := persistencePSP.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()

	// TECHDEBT(#818): consider all staked actors, not just validators.
	validators, err := readCtx.GetAllValidators(int64(height))
	if err != nil {
		return nil, err
	}
	return peerstore_provider.ActorsToPeerstore(persistencePSP, validators)
}

func (persistencePSP *persistencePeerstoreProvider) GetStakedPeerstoreAtCurrentHeight() (typesP2P.Peerstore, error) {
	currentHeight := persistencePSP.GetBus().GetCurrentHeightProvider().CurrentHeight()
	return persistencePSP.GetStakedPeerstoreAtHeight(currentHeight)
}

// GetStakedPeerstoreAtHeight implements the respective `PeerstoreProvider` interface method.
func (persistencePSP *persistencePeerstoreProvider) GetUnstakedPeerstore() (typesP2P.Peerstore, error) {
	// TECHDEBT(#810, #811): use `bus.GetUnstakedActorRouter()` once it's available.
	unstakedActorRouterMod, err := persistencePSP.GetBus().GetModulesRegistry().GetModule(typesP2P.UnstakedActorRouterSubmoduleName)
	if err != nil {
		return nil, err
	}

	unstakedActorRouter, ok := unstakedActorRouterMod.(typesP2P.Router)
	if !ok {
		return nil, fmt.Errorf("unexpected unstaked actor router submodule type: %T", unstakedActorRouterMod)
	}

	return unstakedActorRouter.GetPeerstore(), nil
}
