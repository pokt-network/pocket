package modules

import (
	"pocket/shared/types"
)

type EventsChannel chan types.Event

type Bus interface {
	PublishEventToBus(e *types.Event)
	GetBusEvent() *types.Event
	GetEventBus() EventsChannel

	GetPersistenceModule() PersistenceModule
	GetNetworkModule() NetworkModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
}
