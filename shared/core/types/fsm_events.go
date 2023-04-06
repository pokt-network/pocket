package types

type StateMachineEvent string

const (
	StateMachineEvent_Start StateMachineEvent = "Start"

	StateMachineEvent_P2P_IsBootstrapped StateMachineEvent = "P2P_IsBootstrapped"

	StateMachineEvent_Consensus_IsUnsynced           StateMachineEvent = "Consensus_IsUnsynced"
	StateMachineEvent_Consensus_IsSyncing            StateMachineEvent = "Consensus_IsSyncing"
	StateMachineEvent_Consensus_IsSyncedValidator    StateMachineEvent = "Consensus_IsSyncedValidator"
	StateMachineEvent_Consensus_IsSyncedNonValidator StateMachineEvent = "Consensus_IsSyncedNonValidator"
)
