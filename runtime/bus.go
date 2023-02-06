package runtime

import (
	"log"
	"sync"

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

func (m *bus) RegisterModule(module modules.Module) error {
	module.SetBus(m)
	m.modulesRegistry.RegisterModule(module)
	return nil
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
	var err error
	if mod, err := m.modulesRegistry.GetModule(modules.PersistenceModuleName); err != nil {
		return mod.(modules.PersistenceModule)
	}
	log.Fatalf("%s", err)
	return nil
}

func (m *bus) GetP2PModule() modules.P2PModule {
	var err error
	if mod, err := m.modulesRegistry.GetModule(modules.P2PModuleName); err != nil {
		return mod.(modules.P2PModule)
	}
	log.Fatalf("%s", err)
	return nil
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	var err error
	if mod, err := m.modulesRegistry.GetModule(modules.UtilityModuleName); err != nil {
		return mod.(modules.UtilityModule)
	}
	log.Fatalf("%s", err)
	return nil
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	var err error
	if mod, err := m.modulesRegistry.GetModule(modules.ConsensusModuleName); err != nil {
		return mod.(modules.ConsensusModule)
	}
	log.Fatalf("%s", err)
	return nil
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	var err error
	for _, moduleName := range telemetry.ImplementationNames {
		telemetryMod, err := m.modulesRegistry.GetModule(moduleName)
		if err != nil {
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
	var err error
	if mod, err := m.modulesRegistry.GetModule(modules.LoggerModuleName); err != nil {
		return mod.(modules.LoggerModule)
	}
	log.Fatalf("%s", err)
	return nil
}

func (m *bus) GetRPCModule() modules.RPCModule {
	var err error
	if mod, err := m.modulesRegistry.GetModule(modules.RPCModuleName); err != nil {
		return mod.(modules.RPCModule)
	}
	log.Fatalf("%s", err)
	return nil
}

func (m *bus) GetRuntimeMgr() modules.RuntimeMgr {
	return m.runtimeMgr
}
