package shared

import (
	"log"

	"github.com/pokt-network/pocket/internal/shared/messaging"
	"github.com/pokt-network/pocket/internal/shared/modules"
)

var _ modules.Bus = &bus{}

type bus struct {
	modules.Bus

	// Bus events
	channel modules.EventsChannel

	// Modules
	persistence modules.PersistenceModule
	p2p         modules.P2PModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
	telemetry   modules.TelemetryModule
	logger      modules.LoggerModule
	rpc         modules.RPCModule

	runtimeMgr modules.RuntimeMgr
}

const (
	DefaultPocketBusBufferSize = 100
)

func CreateBus(
	runtimeMgr modules.RuntimeMgr,
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
	telemetry modules.TelemetryModule,
	logger modules.LoggerModule,
	rpc modules.RPCModule,
) (modules.Bus, error) {
	bus := &bus{
		channel: make(modules.EventsChannel, DefaultPocketBusBufferSize),

		runtimeMgr: runtimeMgr,

		persistence: persistence,
		p2p:         p2p,
		utility:     utility,
		consensus:   consensus,
		telemetry:   telemetry,
		logger:      logger,
		rpc:         rpc,
	}

	modules := map[string]modules.Module{
		"persistence": persistence,
		"consensus":   consensus,
		"p2p":         p2p,
		"utility":     utility,
		"telemetry":   telemetry,
		"logger":      logger,
		"rpc":         rpc,
	}

	// checks if modules are not nil and sets their bus to this bus instance.
	// will not carry forward if one of the modules is nil
	for modName, mod := range modules {
		if mod == nil {
			log.Fatalf("Bus Error: the provided %s module is nil, Please use CreateBusWithOptionalModules if you intended it to be nil.", modName)
		}
		mod.SetBus(bus)
	}

	return bus, nil
}

// This is a version of CreateBus that accepts nil modules.
// This function allows you to use a specific module in isolation of other modules by providing a bus with nil modules.
//
// Example of usage: `app/client/main.go`
//
//	We want to use the pre2p module in isolation to communicate with nodes in the network.
//	The pre2p module expects to retrieve a telemetry module through the bus to perform instrumentation, thus we need to inject a bus that can retrieve a telemetry module.
//	However, we don't need telemetry for the dev client.
//	Using `CreateBusWithOptionalModules`, we can create a bus with only pre2p and a NOOP telemetry module
//	so that we can the pre2p module without any issues.
func CreateBusWithOptionalModules(
	runtimeMgr modules.RuntimeMgr,
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
	telemetry modules.TelemetryModule,
	logger modules.LoggerModule,
	rpc modules.RPCModule,
) modules.Bus {
	bus := &bus{
		channel: make(modules.EventsChannel, DefaultPocketBusBufferSize),

		runtimeMgr: runtimeMgr,

		persistence: persistence,
		p2p:         p2p,
		utility:     utility,
		consensus:   consensus,
		telemetry:   telemetry,
		logger:      logger,
		rpc:         rpc,
	}

	maybeSetModuleBus := func(mod modules.Module) {
		if mod != nil {
			mod.SetBus(bus)
		}
	}

	maybeSetModuleBus(persistence)
	maybeSetModuleBus(p2p)
	maybeSetModuleBus(utility)
	maybeSetModuleBus(consensus)
	maybeSetModuleBus(telemetry)
	maybeSetModuleBus(logger)
	maybeSetModuleBus(rpc)

	return bus
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
	return m.persistence
}

func (m *bus) GetP2PModule() modules.P2PModule {
	return m.p2p
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	return m.utility
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	return m.consensus
}

func (m *bus) GetTelemetryModule() modules.TelemetryModule {
	return m.telemetry
}

func (m *bus) GetLoggerModule() modules.LoggerModule {
	return m.logger
}

func (m *bus) GetRPCModule() modules.RPCModule {
	return m.rpc
}

func (m *bus) GetRuntimeMgr() modules.RuntimeMgr {
	return m.runtimeMgr
}
