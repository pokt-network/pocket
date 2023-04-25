package modules

//go:generate mockgen -destination=./mocks/consensus_module_mock.go github.com/pokt-network/pocket/shared/modules ConsensusModule,ConsensusPacemaker,ConsensusStateSync,ConsensusDebugModule

import (
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	ConsensusModuleName      = "consensus"
	PacemakerModuleName      = "pacemaker"
	LeaderElectionModuleName = "leader_election"
)

// NOTE: Consensus is the core of the replicated state machine and is driven by various asynchronous events.
// Consider adding a mutex lock to your implementation that is acquired at the beginning of each entrypoint/function implemented in this interface.
// Make sure that you are not locking again within the same call to avoid deadlocks (for example when the methods below call each other in your implementation).
type ConsensusModule interface {
	Module
	KeyholderModule

	ConsensusStateSync
	ConsensusPacemaker
	ConsensusDebugModule

	// Consensus Engine Handlers
	// TODO: Rename `HandleMessage` to a more specific name that is consistent with its business logic.
	HandleMessage(*anypb.Any) error
	// State Sync message handlers
	HandleStateSyncMessage(*anypb.Any) error
	// FSM transition event handler
	HandleEvent(transitionMessageAny *anypb.Any) error

	// Consensus State Accessors
	CurrentHeight() uint64
	CurrentRound() uint64
	CurrentStep() uint64
}

// ConsensusPacemaker represents functions exposed by the Consensus module for Pacemaker specific business logic.
// These functions are intended to only be called by the Pacemaker module.
// TODO(#428): This interface should be removed when the communication between the pacemaker and consensus module become asynchronous via the bus or go channels.
type ConsensusPacemaker interface {
	// Clearers
	ResetRound(isNewHeight bool)
	// TODO(@deblasis): remove this and implement an event based approach
	ReleaseUtilityUnitOfWork() error

	// Setters
	SetHeight(uint64)
	SetRound(uint64)
	SetStep(uint8) // CONSIDERATION: Change to `typesCons.HotstuffStep; causes an import cycle.

	// Communicators
	BroadcastMessageToValidators(*anypb.Any) error

	// Leader helpers
	IsLeader() bool
	IsLeaderSet() bool
	NewLeader(*anypb.Any) error // CONSIDERATION: Consider changing input to typesCons.HotstuffMessage. This requires to do refactoring.

	// Getters
	IsPrepareQCNil() bool
	GetPrepareQC() (*anypb.Any, error)
	GetNodeId() uint64
}

// ConsensusStateSync exposes functionality of the Consensus module for StateSync specific business logic.
// These functions are intended to only be called by the StateSync module.
// INVESTIGATE: This interface enable a fast implementation of state sync but look into a way of removing it in the future
type ConsensusStateSync interface {
	GetNodeIdFromNodeAddress(string) (uint64, error)
	GetNodeAddress() string
}

// ConsensusDebugModule exposes functionality used for testing & development purposes.
// Not to be used in production.
// TODO: Add a flag so this is not compiled in the prod binary.
type ConsensusDebugModule interface {
	HandleDebugMessage(*messaging.DebugMessage) error

	SetHeight(uint64)
	SetRound(uint64)
	SetStep(uint8) // REFACTOR: This should accept typesCons.HotstuffStep
	SetBlock(*types.Block)

	SetUtilityUnitOfWork(UtilityUnitOfWork)

	// REFACTOR: This should accept typesCons.HotstuffStep and return typesCons.NodeId.
	GetLeaderForView(height, round uint64, step uint8) (leaderId uint64)
}
