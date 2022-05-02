package shared

import (
	"pocket/shared/modules"
	"pocket/shared/types"
)

type bus struct {
	modules.Bus

	channel modules.EventsChannel

	persistence modules.PersistenceModule
	network     modules.NetworkModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
}

const DefaultPocketBusBufferSize = 100 // Create a synchronous event bus by blocking on every message.

func CreateBus(
	channel modules.EventsChannel,
	persistence modules.PersistenceModule,
	network modules.NetworkModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) (modules.Bus, error) {
	// Allow injecting a bus for testing purposes.
	if channel == nil {
		channel = make(modules.EventsChannel, DefaultPocketBusBufferSize)
	}

	bus := &bus{
		channel:     channel,
		persistence: persistence,
		network:     network,
		utility:     utility,
		consensus:   consensus,
	}

	persistence.SetBus(bus)
	consensus.SetBus(bus)
	network.SetBus(bus)
	utility.SetBus(bus)

	return bus, nil
}

func (m *bus) PublishEventToBus(e *types.Event) {
	m.channel <- *e
}

func (m *bus) GetBusEvent() *types.Event {
	e := <-m.channel
	return &e
}

func (m *bus) GetEventBus() modules.EventsChannel {
	return m.channel
}

func (m *bus) GetPersistenceModule() modules.PersistenceModule {
	return m.persistence
}

func (m *bus) GetNetworkModule() modules.NetworkModule {
	return m.network
}

func (m *bus) GetUtilityModule() modules.UtilityModule {
	return m.utility
}

func (m *bus) GetConsensusModule() modules.ConsensusModule {
	return m.consensus
}
