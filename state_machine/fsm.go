package state_machine

import (
	"github.com/looplab/fsm"
)

// NewNodeFSM returns a KISS Finite State Machine that is meant to mimick the various "states" of the node.
//
// source: consensus/doc/PROTOCOL_STATE_SYNC.md + additions for P2P
// The initial implementation is going to be used to understand, from a P2P perspective, if the node requires bootstrapping
func NewNodeFSM(callbacks *fsm.Callbacks, options ...func(*fsm.FSM)) *fsm.FSM {
	var cb = fsm.Callbacks{}
	if callbacks != nil {
		cb = *callbacks
	}

	stateMachine := fsm.NewFSM(
		"stopped",
		fsm.Events{
			{Name: "start", Src: []string{"stopped"}, Dst: "P2P_bootstrapping"},
			{Name: "P2P_isBootstrapped", Src: []string{"P2P_bootstrapping"}, Dst: "P2P_bootstrapped"},
			{Name: "Consensus_isUnsynched", Src: []string{"P2P_bootstrapped"}, Dst: "Consensus_unsynched"},
			{Name: "Consensus_isSyncing", Src: []string{"Consensus_unsynched"}, Dst: "Consensus_syncMode"},
			{Name: "Consensus_isCaughtUp", Src: []string{"P2P_bootstrapped", "Consensus_syncMode"}, Dst: "Consensus_synced"},
		},
		cb,
	)

	for _, option := range options {
		option(stateMachine)
	}

	return stateMachine
}
