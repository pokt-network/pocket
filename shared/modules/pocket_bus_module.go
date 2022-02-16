package modules

import (
	"pocket/shared/types"
)

type PocketEventsChannel chan types.PocketEvent

type BusModule interface {
	PublishEventToBus(e *types.PocketEvent)
	GetBusEvent() *types.PocketEvent
	GetEventBus() PocketEventsChannel

	GetPersistenceModule() PersistenceModule
	GetNetworkModule() NetworkModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
}
