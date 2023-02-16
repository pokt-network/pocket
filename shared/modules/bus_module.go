package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/bus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

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
	RegisterModule(module Module)

	// Pocket modules
	GetPersistenceModule() PersistenceModule
	GetP2PModule() P2PModule
	GetUtilityModule() UtilityModule
	GetConsensusModule() ConsensusModule
	GetTelemetryModule() TelemetryModule
	GetLoggerModule() LoggerModule
	GetRPCModule() RPCModule
	GetStateMachineModule() StateMachineModule

	// Runtime
	GetRuntimeMgr() RuntimeMgr
}
