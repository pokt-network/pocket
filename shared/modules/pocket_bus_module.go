package modules

import (
	"pocket/shared/types"
)

type EventsChannel chan types.PocketEvent

type Bus interface {
	// TODO: Do we want to implement `Module` here as well?

	// Bus Events
	PublishEventToBus(e *types.PocketEvent)
	GetBusEvent() *types.PocketEvent
	GetEventBus() EventsChannel

	// Pocket modules
	GetPersistenceModule() PersistenceModule
	GetP2PModule() P2PModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
}
