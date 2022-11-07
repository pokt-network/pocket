package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	"github.com/pokt-network/pocket/shared/debug"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO(olshansky): deprecate ValidatorMap or populate from persistence module
type ValidatorMap map[string]Actor

// NOTE: Consensus is the core of the replicated state machine and is driven by various asynchronous events.
// Consider adding a mutex lock to your implementation that is acquired at the beginning of each entrypoint/function implemented in this interface.
// Make sure that you are not locking again within the same call to avoid deadlocks (for example when the methods below call each other in your implementation).
type ConsensusModule interface {
	Module
	ConfigurableModule
	GenesisDependentModule
	KeyholderModule

	// Consensus Engine Handlers
	HandleMessage(*anypb.Any) error
	HandleDebugMessage(*debug.DebugMessage) error

	// Consensus State Accessors
	CurrentHeight() uint64
	ValidatorMap() ValidatorMap
}
