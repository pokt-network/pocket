package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

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

	// Consensus engine handlers
	HandleMessage(*anypb.Any) error
	// State Sync message handlers
	HandleStateSyncMessage(*anypb.Any) error

	// Consensus State Accessors
	CurrentHeight() uint64
	CurrentRound() uint64
	CurrentStep() uint64

	// State Sync functions
	EnableServerMode()
}

// ConsensusStateSync represents functions exposed by the Consensus module for StateSync specific business logic.
// These functions are intended to only be called by the StateSync module.
// TODO: This interface enables a fast implementation of state sync but should be removed in the future
type ConsensusStateSync interface {
	GetNodeIdFromNodeAddress(string) (uint64, error)
	GetNodeAddress() string
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
}
