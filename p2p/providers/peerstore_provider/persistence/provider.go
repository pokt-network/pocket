package persistence

import (
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ peerstore_provider.PeerstoreProvider = &persistencePeerstoreProvider{}
	_ persistencePStoreProviderFactory     = &persistencePeerstoreProvider{}
)

type persistencePStoreProviderOption func(*persistencePeerstoreProvider)
type persistencePStoreProviderFactory = modules.FactoryWithOptions[peerstore_provider.PeerstoreProvider, persistencePStoreProviderOption]

// TECHDEBT(#810): refactor to implement `Submodule` interface.
type persistencePeerstoreProvider struct {
	base_modules.IntegratableModule
}

func Create(bus modules.Bus, options ...persistencePStoreProviderOption) (peerstore_provider.PeerstoreProvider, error) {
	return new(persistencePeerstoreProvider).Create(bus, options...)
}

func (*persistencePeerstoreProvider) Create(bus modules.Bus, options ...persistencePStoreProviderOption) (peerstore_provider.PeerstoreProvider, error) {
	pabp := &persistencePeerstoreProvider{
		IntegratableModule: *base_modules.NewIntegratableModule(bus),
	}

	for _, o := range options {
		o(pabp)
	}

	return pabp, nil
}

func (*persistencePeerstoreProvider) GetModuleName() string {
	return peerstore_provider.ModuleName
}

// GetStakedPeerstoreAtHeight implements the respective `PeerstoreProvider` interface method.
func (persistencePSP *persistencePeerstoreProvider) GetStakedPeerstoreAtHeight(height uint64) (typesP2P.Peerstore, error) {
	readCtx, err := persistencePSP.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()

	validators, err := readCtx.GetAllValidators(int64(height))
	if err != nil {
		return nil, err
	}
	return peerstore_provider.ActorsToPeerstore(persistencePSP, validators)
}

// GetStakedPeerstoreAtHeight implements the respective `PeerstoreProvider` interface method.
func (persistencePSP *persistencePeerstoreProvider) GetUnstakedPeerstore() (typesP2P.Peerstore, error) {
	return peerstore_provider.GetUnstakedPeerstore(persistencePSP.GetBus())
}
