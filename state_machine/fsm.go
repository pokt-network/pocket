package state_machine

import (
	"github.com/looplab/fsm"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// NewNodeFSM returns a KISS Finite State Machine that is meant to mimick the various "states" of the node.
//
// The current set of states and events captures a limited subset of state sync and P2P bootstrapping-related events.
// More states & events in any of the modules supported should be added and documented here.
func NewNodeFSM(callbacks *fsm.Callbacks, options ...func(*fsm.FSM)) *fsm.FSM {
	var cb = fsm.Callbacks{}
	if callbacks != nil {
		cb = *callbacks
	}

	stateMachine := fsm.NewFSM(
		string(coreTypes.StateMachineState_Stopped),
		fsm.Events{
			{
				Name: string(coreTypes.StateMachineEvent_Start),
				Src: []string{
					string(coreTypes.StateMachineState_Stopped),
				},
				Dst: string(coreTypes.StateMachineState_P2P_Bootstrapping),
			},
			{
				Name: string(coreTypes.StateMachineEvent_P2P_IsBootstrapped),
				Src: []string{
					string(coreTypes.StateMachineState_P2P_Bootstrapping),
				},
				Dst: string(coreTypes.StateMachineState_P2P_Bootstrapped),
			},
			{
				Name: string(coreTypes.StateMachineEvent_Consensus_IsSyncing),
				Src: []string{
					string(coreTypes.StateMachineState_Consensus_Unsynced),
				},
				Dst: string(coreTypes.StateMachineState_Consensus_SyncMode),
			},
			{
				Name: string(coreTypes.StateMachineEvent_Consensus_IsSyncedValidator),
				Src: []string{
					string(coreTypes.StateMachineState_Consensus_SyncMode),
				},
				Dst: string(coreTypes.StateMachineState_Consensus_Pacemaker),
			},
			{
				Name: string(coreTypes.StateMachineEvent_Consensus_IsSyncedNonValidator),
				Src: []string{
					string(coreTypes.StateMachineState_Consensus_SyncMode),
				},
				Dst: string(coreTypes.StateMachineState_Consensus_Synced),
			},
			{
				Name: string(coreTypes.StateMachineEvent_Consensus_IsUnsynced),
				Src: []string{
					string(coreTypes.StateMachineState_Consensus_Pacemaker),
					string(coreTypes.StateMachineState_Consensus_Synced),
					string(coreTypes.StateMachineState_P2P_Bootstrapped),
				},
				Dst: string(coreTypes.StateMachineState_Consensus_Unsynced),
			},
		},
		cb,
	)

	for _, option := range options {
		option(stateMachine)
	}

	return stateMachine
}
