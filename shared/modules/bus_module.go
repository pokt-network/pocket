package modules

//go:generate mockgen -destination=./mocks/bus_module_mock.go github.com/pokt-network/pocket/shared/modules Bus

import (
	"github.com/pokt-network/pocket/shared/messaging"
)

const BusModuleName = "bus"

// DISCUSS if this channel should be of pointers to PocketEvents or not. Pointers
// would avoid doing object copying, but might also be less thread safe if another goroutine changes
// it, which could potentially be a feature rather than a bug.
type EventsChannel chan *messaging.PocketEnvelope

type Bus interface {
	// Bus Events
	PublishEventToBus(e *messaging.PocketEnvelope)
	GetBusEvent() *messaging.PocketEnvelope
	GetEventBus() EventsChannel

	// Dependency Injection / Service Discovery
	GetModulesRegistry() ModulesRegistry
	RegisterModule(module Submodule)

	// Pocket modules
	GetPersistenceModule() PersistenceModule
	GetP2PModule() P2PModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
	GetTelemetryModule() TelemetryModule
	GetLoggerModule() LoggerModule
	GetRPCModule() RPCModule
	GetStateMachineModule() StateMachineModule
	GetIBCModule() IBCModule

	// Pocket submodules
	GetCurrentHeightProvider() CurrentHeightProvider

	// Runtime
	GetRuntimeMgr() RuntimeMgr
}
