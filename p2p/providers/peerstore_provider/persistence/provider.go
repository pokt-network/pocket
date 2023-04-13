package persistence

import (
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

var _ peerstore_provider.PeerstoreProvider = &persistencePeerstoreProvider{}

type persistencePeerstoreProvider struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule
}

func NewPersistencePeerstoreProvider(bus modules.Bus, options ...func(*persistencePeerstoreProvider)) *persistencePeerstoreProvider {
	pabp := &persistencePeerstoreProvider{
		IntegratableModule: *base_modules.NewIntegratableModule(bus),
	}

	for _, o := range options {
		o(pabp)
	}

	return pabp
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(persistencePeerstoreProvider).Create(bus, options...)
}

func (*persistencePeerstoreProvider) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return NewPersistencePeerstoreProvider(bus), nil
}

func (*persistencePeerstoreProvider) GetModuleName() string {
	return peerstore_provider.ModuleName
}

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

func (pabp *persistencePeerstoreProvider) GetP2PConfig() *configs.P2PConfig {
	return pabp.GetBus().GetRuntimeMgr().GetConfig().P2P
}
