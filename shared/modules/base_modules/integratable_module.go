package base_modules

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.IntegratableModule = &IntegratableModule{}

// IntegratableModule is a base struct that is meant to be embedded in module structs that implement the interface `modules.IntegratableModule`.
//
// It provides the basic logic for the `SetBus` and `GetBus` methods and allows the implementer to reduce boilerplate code keeping the code
// DRY (Don't Repeat Yourself) while preserving the ability to override the methods if needed.
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
