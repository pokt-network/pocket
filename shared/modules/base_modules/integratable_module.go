package base_modules

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.IntegratableModule = &IntegratableModule{}

type IntegratableModule struct {
	bus modules.Bus
}

func NewIntegratableModule(bus modules.Bus) *IntegratableModule {
	return &IntegratableModule{bus: bus}
}

func (m *IntegratableModule) GetBus() modules.Bus {
	return m.bus
}

func (m *IntegratableModule) SetBus(bus modules.Bus) {
	m.bus = bus
}
