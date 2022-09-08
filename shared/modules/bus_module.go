package modules

import (
	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
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

	// Configuration
	GetConfig() *genesis.Config
	GetGenesis() *genesis.GenesisState

	// Time
	GetClock() clock.Clock
}
