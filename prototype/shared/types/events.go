package types

import (
	"google.golang.org/protobuf/types/known/anypb"
	"net"
	"pocket/consensus/types"
)

type SourceModule string
type EventTopic string

const (
	// Core
	CONSENSUS_MODULE SourceModule = "consensus"
	P2P              SourceModule = "p2p"
	persistence      SourceModule = "persistence"
	UTILITY          SourceModule = "utility"

	// Consensus
	STATESYNC       SourceModule = "statesync"
	LEADER_ELECTION SourceModule = "leader_election"

	// Auxiliary
	TEST  SourceModule = "test"
	DEBUG SourceModule = "debug"
)

const (
	// Consensus
	CONSENSUS                   EventTopic = "CONSENSUS"
	CONSENSUS_TELEMETRY_MESSAGE EventTopic = "CONSENSUS_TELEMETRY_MESSAGE"

	// Consensus auxilary
	STATE_SYNC_MESSAGE      EventTopic = "STATE_SYNC_MESSAGE"
	LEADER_ELECTION_MESSAGE EventTopic = "LEADER_ELECTION_MESSAGE"

	// UTILITY?
	UTILITY_TX_MESSAGE       EventTopic = "TRANSACTION_MESSAGE"
	UTILITY_EVIDENCE_MESSAGE EventTopic = "EVIDENCE_MESSAGE"

	// P2P
	P2P_SEND_MESSAGE      EventTopic = "p2p_send_message"
	P2P_BROADCAST_MESSAGE EventTopic = "p2p_broadcast_message"
)

type Event struct {
	SourceModule SourceModule
	Destination  types.NodeId

	// PocketTopic  EventTopic
	// MessageData  []byte
	PocketTopic string
	MessageData *anypb.Any

	NetworkConnection net.Conn // TODO: Only used for debugging/telemetry. Move to PocketContext somehow...
}
