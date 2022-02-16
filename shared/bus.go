package shared

import (
	"pocket/shared/modules"
	"pocket/shared/types"
)

type busModule struct {
	modules.BusModule

	channel modules.PocketEventsChannel

	persistence modules.PersistenceModule
	network     modules.NetworkModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
}

const DefaultPocketBusBufferSize = 100 // Create a synchronous event bus by blocking on every message.

func CreateBus(
	channel modules.PocketEventsChannel,
	persistence modules.PersistenceModule,
	network modules.NetworkModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) (modules.BusModule, error) {
	// Allow injecting a bus for testing purposes.
	if channel == nil {
		channel = make(modules.PocketEventsChannel, DefaultPocketBusBufferSize)
	}

	bus := &busModule{
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

func (m *busModule) PublishEventToBus(e *types.PocketEvent) {
	m.channel <- *e
}

func (m *busModule) GetBusEvent() *types.PocketEvent {
	e := <-m.channel
	return &e
}

func (m *busModule) GetEventBus() modules.PocketEventsChannel {
	return m.channel
}

func (m *busModule) GetPersistenceModule() modules.PersistenceModule {
	return m.persistence
}

func (m *busModule) GetNetworkModule() modules.NetworkModule {
	return m.network
}

func (m *busModule) GetUtilityModule() modules.UtilityModule {
	return m.utility
}

func (m *busModule) GetConsensusModule() modules.ConsensusModule {
	return m.consensus
}

// func (m *busModule) SetpersistenceModule() modules.persistenceModule {
// 	return m.persistence
// }

// func (m *busModule) SetNetworkModule() modules.NetworkModule {
// 	return m.network
// }

// func (m *busModule) SetUtilityModule() modules.UtilityModule {
// 	return m.utility
// }

// func (m *busModule) SetConsensusModule() modules.ConsensusModule {
// 	return m.consensus
// }
