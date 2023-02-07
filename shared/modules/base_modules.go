package modules

var _ InterruptableModule = &BaseInterruptableModule{}

// Noop_interruptable_module is a noop implementation of the InterruptableModule interface.
//
// It is useful for modules that do not need any particular logic to be executed when started or stopped.
// In these situations, just embed this struct into the module struct.
type BaseInterruptableModule struct{}

func (*BaseInterruptableModule) Start() error {
	return nil
}

func (*BaseInterruptableModule) Stop() error {
	return nil
}

var _ IntegratableModule = &BaseIntegratableModule{}

type BaseIntegratableModule struct {
	bus Bus
}

func NewBaseIntegratableModule(bus Bus) *BaseIntegratableModule {
	return &BaseIntegratableModule{bus: bus}
}

func (m *BaseIntegratableModule) GetBus() Bus {
	// if m.bus == nil {
	// 	logger.Global.Fatal().Msg("PocketBus is not initialized")
	// }
	return m.bus
}

func (m *BaseIntegratableModule) SetBus(bus Bus) {
	m.bus = bus
}
