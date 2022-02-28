package modules

import (
	"pocket/shared/types"
)

type EventsChannel chan types.Event

type Bus interface {
	// TODO: Do we want to implement `Module` here as well?

	// Bus Events
	PublishEventToBus(e *types.Event)
	GetBusEvent() *types.Event
	GetEventBus() EventsChannel

	// Pocket modules
	GetPersistenceModule() PersistenceModule
	GetP2PModule() P2PModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
}
