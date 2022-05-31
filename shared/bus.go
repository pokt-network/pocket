package shared

import (
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
)

type bus struct {
	modules.Bus

	channel modules.EventsChannel

	persistence modules.PersistenceModule
	p2p         modules.P2PModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
	telemetry   modules.TelemetryModule
}

const (
	DefaultPocketBusBufferSize = 100
)

func CreateBusWithOptionalModules(
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
	telemetry modules.TelemetryModule,
) modules.Bus {
	bus := &bus{
		channel:     make(modules.EventsChannel, DefaultPocketBusBufferSize),
		persistence: nil,
		p2p:         nil,
		utility:     nil,
		consensus:   nil,
		telemetry:   nil,
	}

	if persistence != nil {
		bus.persistence = persistence
		persistence.SetBus(bus)
	}

	if p2p != nil {
		bus.p2p = p2p
		p2p.SetBus(bus)
	}

	if utility != nil {
		bus.utility = utility
		utility.SetBus(bus)
	}

	if consensus != nil {
		bus.consensus = consensus
		consensus.SetBus(bus)
	}

	if telemetry != nil {
		bus.telemetry = telemetry
		telemetry.SetBus(bus)
	}

	return bus
}

func CreateBus(
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
	telemetry modules.TelemetryModule,
) (modules.Bus, error) {
	bus := &bus{
		channel:     make(modules.EventsChannel, DefaultPocketBusBufferSize),
		persistence: persistence,
		p2p:         p2p,
		utility:     utility,
		consensus:   consensus,
		telemetry:   telemetry,
	}

	persistence.SetBus(bus)
	consensus.SetBus(bus)
	p2p.SetBus(bus)
	utility.SetBus(bus)
	telemetry.SetBus(bus)

	return bus, nil
}

func (m *bus) PublishEventToBus(e *types.PocketEvent) {
	m.channel <- *e
}

func (m *bus) GetBusEvent() *types.PocketEvent {
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
