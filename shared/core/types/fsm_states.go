package types

type StateMachineState string

const (
	StateMachineState_Stopped StateMachineState = "Stopped"

	StateMachineState_P2P_Bootstrapping StateMachineState = "P2P_Bootstrapping"
	StateMachineState_P2P_Bootstrapped  StateMachineState = "P2P_Bootstrapped"

	StateMachineState_Consensus_Unsynced StateMachineState = "Consensus_Unsynced"
	StateMachineState_Consensus_SyncMode StateMachineState = "Consensus_SyncMode"
	StateMachineState_Consensus_Synced   StateMachineState = "Consensus_Synced"

	// Used by synced validators to participate in the HotPOKT lifecycle
	StateMachineState_Consensus_Pacemaker StateMachineState = "Consensus_Pacemaker"
)
