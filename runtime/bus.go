package runtime

import (
	"fmt"
	"sync"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
)

var (
	_                 modules.Bus = &bus{}
	telemetryWarnOnce sync.Once
)

type bus struct {
	modules.Bus

	// Node events
	channel modules.EventsChannel

	modulesRegistry modules.ModulesRegistry

	runtimeMgr modules.RuntimeMgr
}

func CreateBus(runtimeMgr modules.RuntimeMgr, opts ...modules.BusOption) (modules.Bus, error) {
	return new(bus).Create(runtimeMgr, opts...)
}

func (b *bus) Create(runtimeMgr modules.RuntimeMgr, opts ...modules.BusOption) (modules.Bus, error) {
	bus := &bus{
		channel: make(modules.EventsChannel, defaults.DefaultBusBufferSize),

		runtimeMgr:      runtimeMgr,
		modulesRegistry: NewModulesRegistry(),
	}

	for _, o := range opts {
		o(bus)
	}

	return bus, nil
}

func (m *bus) GetModulesRegistry() modules.ModulesRegistry {
	return m.modulesRegistry
}

func (m *bus) RegisterModule(module modules.Module) {
	module.SetBus(m)
	m.modulesRegistry.RegisterModule(module)
}

func (m *bus) PublishEventToBus(e *messaging.PocketEnvelope) {
	fmt.Println("OLSH eventsChannel", m.channel)
	m.channel <- e
}

func (m *bus) GetBusEvent() *messaging.PocketEnvelope {
	e := <-m.channel
	return e
}

func (m *bus) GetEventBus() modules.EventsChannel {
	return m.channel
}

func (m *bus) GetRuntimeMgr() modules.RuntimeMgr {
	return m.runtimeMgr
}

func (m *bus) GetPersistenceModule() modules.PersistenceModule {
	return getModuleFromRegistry[modules.PersistenceModule](m, modules.PersistenceModuleName)
}

func (m *bus) GetP2PModule() modules.P2PModule {
	return getModuleFromRegistry[modules.P2PModule](m, modules.P2PModuleName)
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	return getModuleFromRegistry[modules.UtilityModule](m, modules.UtilityModuleName)
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	return getModuleFromRegistry[modules.ConsensusModule](m, modules.ConsensusModuleName)
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	for _, moduleName := range telemetry.ImplementationNames {
		telemetryMod, err := m.modulesRegistry.GetModule(moduleName)
		if err == nil {
			return telemetryMod.(modules.TelemetryModule)
		}
	}
	telemetryWarnOnce.Do(func() {
		logger.Global.Logger.Warn().
			Str("module", modules.TelemetryModuleName).
			Msg("module not found, creating a default noop module instead")
	})
	// this should happen only if called from the client
	noopModule, err := telemetry.CreateNoopTelemetryModule(m)
	if err != nil {
		logger.Global.Logger.Fatal().
			Err(err).
			Str("module", modules.TelemetryModuleName).
			Msg("failed to create noop telemetry module")
	}
	m.RegisterModule(noopModule)
	return noopModule.(modules.TelemetryModule)
}

func (m *bus) GetLoggerModule() modules.LoggerModule {
	return getModuleFromRegistry[modules.LoggerModule](m, modules.LoggerModuleName)
}

func (m *bus) GetRPCModule() modules.RPCModule {
	return getModuleFromRegistry[modules.RPCModule](m, modules.RPCModuleName)
}

func (m *bus) GetStateMachineModule() modules.StateMachineModule {
	return getModuleFromRegistry[modules.StateMachineModule](m, modules.StateMachineModuleName)
}

// WithEventsChannel is used initialize the bus with a specific events channel
func WithEventsChannel(eventsChannel modules.EventsChannel) modules.BusOption {
	return func(m modules.Bus) {
		if m, ok := m.(*bus); ok {
			m.channel = eventsChannel
		}
	}
}

// getModuleFromRegistry is a helper function to get a module from the registry that handles errors and casting via generics
func getModuleFromRegistry[T modules.Module](m *bus, moduleName string) T {
	mod, err := m.modulesRegistry.GetModule(moduleName)
	if err != nil {
		logger.Global.Logger.Fatal().
			Err(err).
			Str("module", moduleName).
			Msg("failed to get module from modulesRegistry")
	}
	return mod.(T)
}
