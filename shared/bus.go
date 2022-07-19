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
}

const (
	DefaultPocketBusBufferSize = 100
)

func CreateBus(
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) (modules.Bus, error) {
	bus := &bus{
		channel:     make(modules.EventsChannel, DefaultPocketBusBufferSize),
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

func CreateBusWithOptionalModules(
	persistence modules.PersistenceModule,
	p2p modules.P2PModule,
	utility modules.UtilityModule,
	consensus modules.ConsensusModule,
) modules.Bus {
	bus := &bus{
		channel:     make(modules.EventsChannel, DefaultPocketBusBufferSize),
		persistence: persistence,
		p2p:         p2p,
		utility:     utility,
		consensus:   consensus,
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

	return bus
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
