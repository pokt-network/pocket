package modules

import "pocket/consensus/pkg/shared/events"

type PocketBus chan events.PocketEvent

type PocketBusModule interface {
	PublishEventToBus(e *events.PocketEvent)
	GetBusEvent() *events.PocketEvent
	GetEventBus() PocketBus

	GetPersistanceModule() PersistanceModule
	GetNetworkModule() NetworkModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
}
