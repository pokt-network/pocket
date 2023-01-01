package runtime

import (
	"log"

	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
)

var _ modules.Bus = &bus{}

type bus struct {
	modules.Bus

	// Bus events
	channel modules.EventsChannel

	// Modules
	modulesMap map[string]modules.Module

	runtimeMgr modules.RuntimeMgr
}

const (
	DefaultPocketBusBufferSize = 100
)

func CreateBus(runtimeMgr modules.RuntimeMgr) (modules.Bus, error) {
	return new(bus).Create(runtimeMgr)
}

func (b *bus) Create(runtimeMgr modules.RuntimeMgr) (modules.Bus, error) {
	bus := &bus{
		channel: make(modules.EventsChannel, DefaultPocketBusBufferSize),

		runtimeMgr: runtimeMgr,
		modulesMap: make(map[string]modules.Module),
	}

	return bus, nil
}

func (m *bus) RegisterModule(module modules.Module) error {
	m.modulesMap[module.GetModuleName()] = module
	m.modulesMap[module.GetModuleName()].SetBus(m)
	return nil
}

func (m *bus) PublishEventToBus(e *messaging.PocketEnvelope) {
	m.channel <- *e
}

func (m *bus) GetBusEvent() *messaging.PocketEnvelope {
	e := <-m.channel
	return &e
}

func (m *bus) GetEventBus() modules.EventsChannel {
	return m.channel
}

func (m *bus) GetPersistenceModule() modules.PersistenceModule {
	return m.modulesMap[modules.PersistenceModuleName].(modules.PersistenceModule)
}

func (m *bus) GetP2PModule() modules.P2PModule {
	return m.modulesMap[modules.P2PModuleName].(modules.P2PModule)
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	return m.modulesMap[modules.UtilityModuleName].(modules.UtilityModule)
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	return m.modulesMap[modules.ConsensusModuleName].(modules.ConsensusModule)
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	telemetryModules := []string{modules.TelemetryModuleName, "telemetry_prometheus", "telemetry_noOP"}
	for _, moduleName := range telemetryModules {
		telemetryMod, ok := m.modulesMap[moduleName]
		if ok {
			return telemetryMod.(modules.TelemetryModule)
		}
	}
	log.Printf("[WARNING] telemetry module not found, using noop telemetry module instead")
	// this should happen only if called from the client
	noopModule, err := telemetry.CreateNoopTelemetryModule(m)
	if err != nil {
		log.Fatalf("failed to create noop telemetry module: %v", err)
	}
	m.RegisterModule(noopModule)
	return noopModule.(modules.TelemetryModule)
}

func (m *bus) GetLoggerModule() modules.LoggerModule {
	return m.modulesMap[modules.LoggerModuleName].(modules.LoggerModule)
}

func (m *bus) GetRPCModule() modules.RPCModule {
	return m.modulesMap[modules.RPCModuleName].(modules.RPCModule)
}

func (m *bus) GetRuntimeMgr() modules.RuntimeMgr {
	return m.runtimeMgr
}
