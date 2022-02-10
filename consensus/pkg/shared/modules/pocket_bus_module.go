package modules

import (
	"pocket/shared/events"
	"pocket/shared/modules"
)

type PocketBus chan events.PocketEvent

type PocketBusModule interface {
	PublishEventToBus(e *events.PocketEvent)
	GetBusEvent() *events.PocketEvent
	GetEventBus() PocketBus

	GetpersistenceModule() modules.PersistenceModule
	GetNetworkModule() modules.NetworkModule
	GetUtilityModule() modules.UtilityModule
	GetConsensusModule() modules.ConsensusModule
}
