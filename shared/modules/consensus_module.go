package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/consensus_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
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
	ConsensusPacemaker

	// Consensus Engine Handlers
	HandleMessage(*anypb.Any) error
	// TODO(gokhan): move it into a debug module
	HandleDebugMessage(*messaging.DebugMessage) error

	// Consensus State Accessors
	CurrentHeight() uint64
	CurrentRound() uint64
	CurrentStep() uint64
}

// This interface represents functions exposed by the Consensus module for Pacemaker specific business logic.
// These functions are intended to only be called by the Pacemaker module.
// TODO(#428): This interface will be removed when the communication between the pacemaker and consensus module become asynchronous.
type ConsensusPacemaker interface {

	//Pacemaker Consensus interaction modules
	ResetRound()
	ClearLeaderMessagesPool()
	SetHeight(uint64)
	SetRound(uint64)

	// IMPROVE: Consider changing the input to type to `typesCons.HotstuffStep`. This currently causes an import cycle and requires significant refactoring.
	SetStep(uint8)
	ResetForNewHeight()
	ReleaseUtilityContext() error
	BroadcastMessageToValidators(*anypb.Any) error
	IsLeader() bool
	IsLeaderSet() bool
	//IMPROVE: Consider changing input to typesCons.HotstuffMessage. This requires to do refactoring.
	NewLeader(*anypb.Any) error
	GetPrepareQC() *anypb.Any
	GetNodeId() uint64
}
