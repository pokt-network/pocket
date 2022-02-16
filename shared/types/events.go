package types

import (
	"google.golang.org/protobuf/types/known/anypb"
	"net"
	"pocket/consensus/types"
)

type SourceModule string
type PocketEventTopic string

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
	CONSENSUS                   PocketEventTopic = "CONSENSUS"
	CONSENSUS_TELEMETRY_MESSAGE PocketEventTopic = "CONSENSUS_TELEMETRY_MESSAGE"

	// Consensus auxilary
	STATE_SYNC_MESSAGE      PocketEventTopic = "STATE_SYNC_MESSAGE"
	LEADER_ELECTION_MESSAGE PocketEventTopic = "LEADER_ELECTION_MESSAGE"

	// UTILITY?
	UTILITY_TX_MESSAGE       PocketEventTopic = "TRANSACTION_MESSAGE"
	UTILITY_EVIDENCE_MESSAGE PocketEventTopic = "EVIDENCE_MESSAGE"

	// P2P
	P2P_SEND_MESSAGE      PocketEventTopic = "p2p_send_message"
	P2P_BROADCAST_MESSAGE PocketEventTopic = "p2p_broadcast_message"
)

type PocketEvent struct {
	SourceModule SourceModule
	Destination  types.NodeId

	// PocketTopic  PocketEventTopic
	// MessageData  []byte
	PocketTopic string
	MessageData *anypb.Any

	NetworkConnection net.Conn // TODO: Only used for debugging/telemetry. Move to PocketContext somehow...
}
