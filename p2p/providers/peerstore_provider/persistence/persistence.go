package persistence

import (
	"fmt"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ peerstore_provider.PeerstoreProvider = &persistencePeerstoreProvider{}

type persistencePStoreProviderOption func(*persistencePeerstoreProvider)
type persistencePStoreProviderFactory = modules.FactoryWithOptions[peerstore_provider.PeerstoreProvider, persistencePStoreProviderOption]

type persistencePeerstoreProvider struct {
	base_modules.IntegratableModule
	persistencePStoreProviderFactory
}

// unstakedPeerstoreProvider is an interface which the p2p module supports in
// order to allow access to the unstaked-actor-router's peerstore
//
// TECHDEBT(#xxx): will become unnecessary after `modules.P2PModule#GetUnstakedPeerstore` is added.`
// CONSIDERATION: split `PeerstoreProvider` into `StakedPeerstoreProvider` and `UnstakedPeerstoreProvider`.
// (see: https://github.com/pokt-network/pocket/pull/804#issuecomment-1576531916)
type unstakedPeerstoreProvider interface {
	GetUnstakedPeerstore() (typesP2P.Peerstore, error)
}

func NewPersistencePeerstoreProvider(bus modules.Bus, options ...persistencePStoreProviderOption) (peerstore_provider.PeerstoreProvider, error) {
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
func (pabp *persistencePeerstoreProvider) GetStakedPeerstoreAtHeight(height uint64) (typesP2P.Peerstore, error) {
	readCtx, err := pabp.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer readCtx.Release()

	validators, err := readCtx.GetAllValidators(int64(height))
	if err != nil {
		return nil, err
	}
	return peerstore_provider.ActorsToPeerstore(pabp, validators)
}

// GetStakedPeerstoreAtHeight implements the respective `PeerstoreProvider` interface method.
func (pabp *persistencePeerstoreProvider) GetUnstakedPeerstore() (typesP2P.Peerstore, error) {
	p2pModule := pabp.GetBus().GetP2PModule()
	if p2pModule == nil {
		return nil, fmt.Errorf("p2p module is not registered to bus and is required")
	}

	unstakedPSP, ok := p2pModule.(unstakedPeerstoreProvider)
	if !ok {
		return nil, fmt.Errorf("p2p module does not implement unstakedPeerstoreProvider")
	}
	return unstakedPSP.GetUnstakedPeerstore()
}
