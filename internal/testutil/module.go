package testutil

import (
	"github.com/foxcpp/go-mockdns"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
)

// TODO_THIS_COMMIT: is this helpful?
const TestModuleName = "testModule"

var (
	_ modules.Module                   = &TestModule{}
	_ modules.ModuleFactoryWithOptions = &TestModule{}
)

type TestModule struct {
	base_modules.IntegratableModule
	base_modules.InterruptableModule

	DNS               *mockdns.Server
	Genesis           *genesis.GenesisState
	Libp2pNetworkMock mocknet.Mocknet
}

func (m *TestModule) GetModuleName() string {
	return TestModuleName
}

func (m *TestModule) Create(
	bus modules.Bus,
	opts ...modules.ModuleOption,
) (modules.Module, error) {
	panic("implement me")
}

func (m *TestModule) GetDNS() *mockdns.Server {
	return m.DNS
}
