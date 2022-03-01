package modules

import (
	"github.com/pokt-network/pocket/shared/types"
)

type EventsChannel chan types.PocketEvent

type Bus interface {
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
