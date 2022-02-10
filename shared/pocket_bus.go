package shared

import (
	"pocket/shared/events"
	"pocket/shared/modules"
)

type pocketBusModule struct {
	modules.PocketBusModule

	channel modules.PocketEventsChannel

	persistence modules.PersistenceModule
	network     modules.NetworkModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
}

const DefaultPocketBusBufferSize = 100 // Create a synchronous event bus by blocking on every message.

func CreatePocketBus(
	channel modules.PocketEventsChannel,
	persistence modules.PersistenceModule,
	network modules.NetworkModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) (modules.PocketBusModule, error) {
	// Allow injecting a bus for testing purposes.
	if channel == nil {
		channel = make(modules.PocketEventsChannel, DefaultPocketBusBufferSize)
	}

	bus := &pocketBusModule{
		channel:     channel,
		persistence: persistence,
		network:     network,
		utility:     utility,
		consensus:   consensus,
	}

	persistence.SetPocketBusMod(bus)
	consensus.SetPocketBusMod(bus)
	network.SetPocketBusMod(bus)
	utility.SetPocketBusMod(bus)

	return bus, nil
}

func (m *pocketBusModule) PublishEventToBus(e *events.PocketEvent) {
	m.channel <- *e
}

func (m *pocketBusModule) GetBusEvent() *events.PocketEvent {
	e := <-m.channel
	return &e
}

func (m *pocketBusModule) GetEventBus() modules.PocketEventsChannel {
	return m.channel
}

func (m *pocketBusModule) GetPersistenceModule() modules.PersistenceModule {
	return m.persistence
}

func (m *pocketBusModule) GetNetworkModule() modules.NetworkModule {
	return m.network
}

func (m *pocketBusModule) GetUtilityModule() modules.UtilityModule {
	return m.utility
}

func (m *pocketBusModule) GetConsensusModule() modules.ConsensusModule {
	return m.consensus
}

// func (m *pocketBusModule) SetpersistenceModule() modules.persistenceModule {
// 	return m.persistence
// }

// func (m *pocketBusModule) SetNetworkModule() modules.NetworkModule {
// 	return m.network
// }

// func (m *pocketBusModule) SetUtilityModule() modules.UtilityModule {
// 	return m.utility
// }

// func (m *pocketBusModule) SetConsensusModule() modules.ConsensusModule {
// 	return m.consensus
// }
