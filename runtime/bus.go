package runtime

import (
	"log"
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

func CreateBus(runtimeMgr modules.RuntimeMgr) (modules.Bus, error) {
	return new(bus).Create(runtimeMgr)
}

func (b *bus) Create(runtimeMgr modules.RuntimeMgr) (modules.Bus, error) {
	bus := &bus{
		channel: make(modules.EventsChannel, defaults.DefaultBusBufferSize),

		runtimeMgr:      runtimeMgr,
		modulesRegistry: NewModulesRegistry(),
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
	m.channel <- e
}

func (m *bus) GetBusEvent() *messaging.PocketEnvelope {
	e := <-m.channel
	return e
}

func (m *bus) GetEventBus() modules.EventsChannel {
	return m.channel
}

func (m *bus) GetPersistenceModule() modules.PersistenceModule {
	mod, err := m.modulesRegistry.GetModule(modules.PersistenceModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.PersistenceModule)
}

func (m *bus) GetP2PModule() modules.P2PModule {
	mod, err := m.modulesRegistry.GetModule(modules.P2PModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.P2PModule)
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	mod, err := m.modulesRegistry.GetModule(modules.UtilityModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.UtilityModule)
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	mod, err := m.modulesRegistry.GetModule(modules.ConsensusModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.ConsensusModule)
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	for _, moduleName := range telemetry.ImplementationNames {
		telemetryMod, err := m.modulesRegistry.GetModule(moduleName)
		if err == nil {
			return telemetryMod.(modules.TelemetryModule)
		}
	}
	telemetryWarnOnce.Do(func() {
		log.Printf("[WARNING] telemetry module not found, creating a default noop telemetry module instead")
	})
	// this should happen only if called from the client
	noopModule, err := telemetry.CreateNoopTelemetryModule(m)
	if err != nil {
		log.Fatalf("failed to create noop telemetry module: %v", err)
	}
	m.RegisterModule(noopModule)
	return noopModule.(modules.TelemetryModule)
}

func (m *bus) GetLoggerModule() modules.LoggerModule {
	mod, err := m.modulesRegistry.GetModule(modules.LoggerModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.LoggerModule)
}

func (m *bus) GetRPCModule() modules.RPCModule {
	mod, err := m.modulesRegistry.GetModule(modules.RPCModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.RPCModule)
}

func (m *bus) GetRuntimeMgr() modules.RuntimeMgr {
	return m.runtimeMgr
}

func (m *bus) GetStateMachineModule() modules.StateMachineModule {
	mod, err := m.modulesRegistry.GetModule(modules.StateMachineModuleName)
	if err != nil {
		logger.Global.Logger.Fatal().Err(err).Msg("failed to get module from modulesRegistry")
	}
	return mod.(modules.StateMachineModule)
}
