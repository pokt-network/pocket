package types

import (
	"google.golang.org/protobuf/types/known/anypb"
)

type SourceModule string
type EventTopic string

const (
	// Core
	CONSENSUS   SourceModule = "consensus"
	P2P         SourceModule = "p2p"
	PERSISTENCE SourceModule = "persistence"
	UTILITY     SourceModule = "utility"

	// Consensus
	STATESYNC       SourceModule = "statesync"
	LEADER_ELECTION SourceModule = "leader_election"

	// Auxiliary
	TEST  SourceModule = "test"
	DEBUG SourceModule = "debug"
)

const (
	// Consensus
	CONSENSUS_MESSAGE           EventTopic = "CONSENSUS_MESSAGE"
	CONSENSUS_TELEMETRY_MESSAGE EventTopic = "CONSENSUS_TELEMETRY_MESSAGE"

	// Consensus auxilary
	STATE_SYNC_MESSAGE      EventTopic = "STATE_SYNC_MESSAGE"
	LEADER_ELECTION_MESSAGE EventTopic = "LEADER_ELECTION_MESSAGE"

	// Utility
	UTILITY_TX_MESSAGE       EventTopic = "TRANSACTION_MESSAGE"
	UTILITY_EVIDENCE_MESSAGE EventTopic = "EVIDENCE_MESSAGE"

	// P2P
	P2P_SEND_MESSAGE      EventTopic = "p2p_send_message"
	P2P_BROADCAST_MESSAGE EventTopic = "p2p_broadcast_message"
)

type Event struct {
	SourceModule SourceModule
	// TODO(olshansky): Either add this back in or remove it altogether.
	// Destination  types.NodeId

	PocketTopic string
	MessageData *anypb.Any
}
