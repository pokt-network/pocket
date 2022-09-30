package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/bus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"encoding/json"

	"github.com/pokt-network/pocket/shared/debug"
)

// DISCUSS if this channel should be of pointers to PocketEvents or not. Pointers
// would avoid doing object copying, but might also be less thread safe if another goroutine changes
// it, which could potentially be a feature rather than a bug.
type EventsChannel chan debug.PocketEvent

type Bus interface {
	// Bus Events
	PublishEventToBus(e *debug.PocketEvent)
	GetBusEvent() *debug.PocketEvent
	GetEventBus() EventsChannel

	// Pocket modules
	GetPersistenceModule() PersistenceModule
	GetP2PModule() P2PModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
	GetTelemetryModule() TelemetryModule

	// Configuration
	GetConfig() map[string]json.RawMessage
	GetGenesis() map[string]json.RawMessage
}
