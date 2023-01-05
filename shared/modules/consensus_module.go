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

// This interface represents functions built for an intermediate solution towards seperation consensus and pacemaker modules
// This functions should be only called by the PaceMaker module.
type ConsensusPacemaker interface {
	//Pacemaker Consensus interaction modules
	ClearLeaderMessagesPool()
	SetHeight(uint64)
	SetRound(uint64)
	SetStep(uint64)
	ResetForNewHeight()
	ReleaseUtilityContext() error
	BroadcastMessageToNodes(*anypb.Any) error
	IsLeader() bool
	IsLeaderSet() bool
	NewLeader(*anypb.Any) error
	GetPrepareQC() *anypb.Any
}
