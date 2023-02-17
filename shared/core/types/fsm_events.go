package types

type StateMachineEvent string

const (
	StateMachineEvent_Start StateMachineEvent = "Start"

	StateMachineEvent_P2P_IsBootstrapped StateMachineEvent = "P2P_IsBootstrapped"

	StateMachineEvent_Consensus_IsUnsynched StateMachineEvent = "Consensus_IsUnsynched"
	StateMachineEvent_Consensus_IsSyncing   StateMachineEvent = "Consensus_IsSyncing"
	StateMachineEvent_Consensus_IsCaughtUp  StateMachineEvent = "Consensus_IsCaughtUp"
)
