package types

type StateMachineEvent string

const (
	StateMachineEvent_Start StateMachineEvent = "Start"

	StateMachineEvent_P2P_IsBootstrapped StateMachineEvent = "P2P_IsBootstrapped"

	StateMachineEvent_Consensus_IsUnsynced            StateMachineEvent = "Consensus_IsUnsynced"
	StateMachineEvent_Consensus_IsSyncing             StateMachineEvent = "Consensus_IsSyncing"
	StateMachineEvent_Consensus_IsSynchedValidator    StateMachineEvent = "Consensus_IsSynchedValidator"
	StateMachineEvent_Consensus_IsSynchedNonValidator StateMachineEvent = "Consensus_IsSynchedNonValidator"
)
