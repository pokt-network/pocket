package modules

import (
	"github.com/pokt-network/pocket/shared/types"
)

// TODO(design): Discuss if this channel should be of pointers to PocketEvents or not. Pointers
// would avoid doing object copying, but might also be less thread safe if another goroutine changes
// it, which could potentially be a feature rather than a bug.
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
	GetTelemetryModule() TelemetryModule
}
