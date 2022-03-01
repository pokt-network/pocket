package modules

import (
	"github.com/pokt-network/pocket/shared/types"
)

type EventsChannel chan types.Event

type Bus interface {
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
