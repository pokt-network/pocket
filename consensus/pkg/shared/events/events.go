package events

import (
	"net"

	"pocket/consensus/pkg/types"
)

type SourceModule string
type PocketEventTopic string

const (
	// Core
	CONSENSUS   SourceModule = "consensus"
	P2P         SourceModule = "p2p"
	PERSISTANCE SourceModule = "persistance"
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
	CONSENSUS_MESSAGE           PocketEventTopic = "CONSENSUS_MESSAGE"
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
	PocketTopic  PocketEventTopic
	MessageData  []byte
	Destination  types.NodeId

	NetworkConnection net.Conn // TODO: Only used for debugging/telemetry. Move to PocketContext somehow...
}
