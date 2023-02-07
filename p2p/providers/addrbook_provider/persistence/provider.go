package persistence

import (
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/transport"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ addrbook_provider.AddrBookProvider = &persistenceAddrBookProvider{}

type persistenceAddrBookProvider struct {
	modules.BaseIntegratableModule
	modules.BaseInterruptableModule

	connFactory typesP2P.ConnectionFactory
}

func NewPersistenceAddrBookProvider(bus modules.Bus, options ...func(*persistenceAddrBookProvider)) *persistenceAddrBookProvider {
	pabp := &persistenceAddrBookProvider{
		BaseIntegratableModule: *modules.NewBaseIntegratableModule(bus),
		connFactory:            transport.CreateDialer, // default connection factory, overridable with WithConnectionFactory()
	}

	for _, o := range options {
		o(pabp)
	}

	return pabp
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(persistenceAddrBookProvider).Create(bus, options...)
}

func (*persistenceAddrBookProvider) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return NewPersistenceAddrBookProvider(bus), nil
}

func (*persistenceAddrBookProvider) GetModuleName() string {
	return addrbook_provider.ModuleName
}

func (pabp *persistenceAddrBookProvider) GetStakedAddrBookAtHeight(height uint64) (typesP2P.AddrBook, error) {
	persistenceReadContext, err := pabp.GetBus().GetPersistenceModule().NewReadContext(int64(height))
	if err != nil {
		return nil, err
	}
	defer persistenceReadContext.Close()

	validators, err := persistenceReadContext.GetAllValidators(int64(height))
	if err != nil {
		return nil, err
	}
	return addrbook_provider.ActorsToAddrBook(pabp, validators)
}

func (pabp *persistenceAddrBookProvider) GetConnFactory() typesP2P.ConnectionFactory {
	return pabp.connFactory
}

func (pabp *persistenceAddrBookProvider) GetP2PConfig() *configs.P2PConfig {
	return pabp.GetBus().GetRuntimeMgr().GetConfig().P2P
}

func (pabp *persistenceAddrBookProvider) SetConnectionFactory(connFactory typesP2P.ConnectionFactory) {
	pabp.connFactory = connFactory
}
