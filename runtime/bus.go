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

	modulesMap map[string]modules.Module

	runtimeMgr modules.RuntimeMgr
}

func CreateBus(runtimeMgr modules.RuntimeMgr) (modules.Bus, error) {
	return new(bus).Create(runtimeMgr)
}

func (b *bus) Create(runtimeMgr modules.RuntimeMgr) (modules.Bus, error) {
	bus := &bus{
		channel: make(modules.EventsChannel, defaults.DefaultBusBufferSize),

		runtimeMgr: runtimeMgr,
		modulesMap: make(map[string]modules.Module),
	}

	return bus, nil
}

func (m *bus) RegisterModule(module modules.Module) error {
	module.SetBus(m)
	m.modulesMap[module.GetModuleName()] = module
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
	if mod, ok := m.modulesMap[modules.PersistenceModuleName]; ok {
		return mod.(modules.PersistenceModule)
	}
	log.Fatalf("%s", ErrModuleNotRegistered("persistence"))
	return nil
}

func (m *bus) GetP2PModule() modules.P2PModule {
	if mod, ok := m.modulesMap[modules.P2PModuleName]; ok {
		return mod.(modules.P2PModule)
	}
	log.Fatalf("%s", ErrModuleNotRegistered("P2P"))
	return nil
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	if mod, ok := m.modulesMap[modules.UtilityModuleName]; ok {
		return mod.(modules.UtilityModule)
	}
	log.Fatalf("%s", ErrModuleNotRegistered(modules.UtilityModuleName))
	return nil
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	if mod, ok := m.modulesMap[modules.ConsensusModuleName]; ok {
		return mod.(modules.ConsensusModule)
	}
	log.Fatalf("%s", ErrModuleNotRegistered(modules.ConsensusModuleName))
	return nil
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	for _, moduleName := range telemetry.ImplementationNames {
		telemetryMod, ok := m.modulesMap[moduleName]
		if ok {
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
	if err := m.RegisterModule(noopModule); err != nil {
		log.Fatalf("[ERROR] Failed to register telemetry module: %v", err.Error())
	}
	return noopModule.(modules.TelemetryModule)
}

func (m *bus) GetLoggerModule() modules.LoggerModule {
	if mod, ok := m.modulesMap[modules.LoggerModuleName]; ok {
		return mod.(modules.LoggerModule)
	}
	log.Fatalf("%s", ErrModuleNotRegistered(modules.LoggerModuleName))
	return nil
}

func (m *bus) GetRPCModule() modules.RPCModule {
	if mod, ok := m.modulesMap[modules.RPCModuleName]; ok {
		return mod.(modules.RPCModule)
	}
	log.Fatalf("%s", ErrModuleNotRegistered(modules.RPCModuleName))
	return nil
}

func (m *bus) GetRuntimeMgr() modules.RuntimeMgr {
	return m.runtimeMgr
}
