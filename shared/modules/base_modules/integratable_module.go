package base_modules

import "github.com/pokt-network/pocket/shared/modules"

var _ modules.IntegrableModule = &IntegrableModule{}

// IntegrableModule is a base struct that is meant to be embedded in module structs that implement the interface `modules.IntegrableModule`.
//
// It provides the basic logic for the `SetBus` and `GetBus` methods and allows the implementer to reduce boilerplate code keeping the code
// DRY (Don't Repeat Yourself) while preserving the ability to override the methods if needed.
type IntegrableModule struct {
	bus modules.Bus
}

func NewIntegrableModule(bus modules.Bus) *IntegrableModule {
	return &IntegrableModule{bus: bus}
}

func (m *IntegrableModule) GetBus() modules.Bus {
	return m.bus
}

func (m *IntegrableModule) SetBus(bus modules.Bus) {
	m.bus = bus
}
