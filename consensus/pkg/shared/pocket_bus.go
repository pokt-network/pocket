package shared

import (
	"pocket/consensus/pkg/shared/events"
	"pocket/consensus/pkg/shared/modules"
)

type pocketBusModule struct {
	modules.PocketBusModule

	pocketBus modules.PocketBus

	persistance modules.PersistanceModule
	network     modules.NetworkModule
	utility     modules.UtilityModule
	consensus   modules.ConsensusModule
}

const DefaultPocketBusBufferSize = 100 // Create a synchronous event bus by blocking on every message.

func CreatePocketBus(
	bus modules.PocketBus,
	persistance modules.PersistanceModule,
	network modules.NetworkModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) (modules.PocketBusModule, error) {
	// Allow injecting a bus for testing purposes.
	if bus == nil {
		bus = make(modules.PocketBus, DefaultPocketBusBufferSize)
	}
	return &pocketBusModule{
		pocketBus:   bus,
		persistance: persistance,
		network:     network,
		utility:     utility,
		consensus:   consensus,
	}, nil
}

func (m *pocketBusModule) PublishEventToBus(e *events.PocketEvent) {
	m.pocketBus <- *e
}

func (m *pocketBusModule) GetBusEvent() *events.PocketEvent {
	e := <-m.pocketBus
	return &e
}

func (m *pocketBusModule) GetEventBus() modules.PocketBus {
	return m.pocketBus
}

func (m *pocketBusModule) GetPersistanceModule() modules.PersistanceModule {
	return m.persistance
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
