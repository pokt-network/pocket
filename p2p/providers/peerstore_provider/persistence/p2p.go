package persistence

import (
	"fmt"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var (
	_ peerstore_provider.PeerstoreProvider = &p2pPeerstoreProvider{}
	_ p2pPStoreProviderFactory             = &p2pPeerstoreProvider{}
)

// unstakedPeerstoreProvider is an interface which the p2p module supports in
// order to allow access to the unstaked-actor-router's peerstore
//
// TECHDEBT(#xxx): will become unnecessary after `modules.P2PModule#GetUnstakedPeerstore` is added.`
// CONSIDERATION: split `PeerstoreProvider` into `StakedPeerstoreProvider` and `UnstakedPeerstoreProvider`.
// (see: https://github.com/pokt-network/pocket/pull/804#issuecomment-1576531916)
type unstakedPeerstoreProvider interface {
	GetUnstakedPeerstore() (typesP2P.Peerstore, error)
}

type p2pPStoreProviderFactory = modules.Factory[peerstore_provider.PeerstoreProvider]

type p2pPeerstoreProvider struct {
	base_modules.IntegratableModule
	persistencePeerstoreProvider

	p2pModule modules.P2PModule
}

func NewP2PPeerstoreProvider(
	bus modules.Bus,
) (peerstore_provider.PeerstoreProvider, error) {
	return new(p2pPeerstoreProvider).Create(bus)
}

func (*p2pPeerstoreProvider) Create(
	bus modules.Bus,
) (peerstore_provider.PeerstoreProvider, error) {
	if bus == nil {
		return nil, fmt.Errorf("bus is required")
	}

	p2pModule := bus.GetP2PModule()
	if p2pModule == nil {
		return nil, fmt.Errorf("p2p module is not registered to bus and is required")
	}

	p2pPSP := &p2pPeerstoreProvider{
		IntegratableModule: *base_modules.NewIntegratableModule(bus),
		p2pModule:          p2pModule,
	}

	return p2pPSP, nil
}

func (*p2pPeerstoreProvider) GetModuleName() string {
	return peerstore_provider.ModuleName
}

func (p2pPSP *p2pPeerstoreProvider) GetUnstakedPeerstore() (typesP2P.Peerstore, error) {
	unstakedPSP, ok := p2pPSP.p2pModule.(unstakedPeerstoreProvider)
	if !ok {
		return nil, fmt.Errorf("p2p module does not implement unstakedPeerstoreProvider")
	}
	return unstakedPSP.GetUnstakedPeerstore()
}
