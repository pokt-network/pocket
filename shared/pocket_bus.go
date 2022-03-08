package shared

import (
	"log"

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
}

const (
	DefaultPocketBusBufferSize = 100
)

func CreateBus(
	channel modules.EventsChannel, // A channel can be injected into the bus only for testing purposes
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) (modules.Bus, error) {
	if channel == nil {
		log.Print("Creating a new Go channel for the Pocket bus...")
		channel = make(modules.EventsChannel, DefaultPocketBusBufferSize)
	}

	bus := &bus{
		channel:     channel,
		persistence: persistence,
		p2p:         p2p,
		utility:     utility,
		consensus:   consensus,
	}

	persistence.SetBus(bus)
	consensus.SetBus(bus)
	p2p.SetBus(bus)
	utility.SetBus(bus)

	return bus, nil
}

func (m *bus) PublishEventToBus(e *types.PocketEvent) {
	m.channel <- e
}

func (m *bus) GetBusEvent() *types.PocketEvent {
	e := <-m.channel
	return e
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
