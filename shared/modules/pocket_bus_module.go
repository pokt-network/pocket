package modules

import (
	// "pocket/consensus/pkg/persistence"

	"pocket/shared/events"
	// "pocket/utility/utility"
)

type PocketEventsChannel chan events.PocketEvent

type PocketBusModule interface {
	PublishEventToBus(e *events.PocketEvent)
	GetBusEvent() *events.PocketEvent
	GetEventBus() PocketEventsChannel

	GetPersistenceModule() PersistenceModule
	GetNetworkModule() NetworkModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
}
