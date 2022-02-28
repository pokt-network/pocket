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

	// Auxiliary
	TEST  SourceModule = "test"
	DEBUG SourceModule = "debug"
)

const (
	// Consensus
	CONSENSUS_MESSAGE EventTopic = "CONSENSUS_MESSAGE"
)

type Event struct {
	SourceModule SourceModule
	PocketTopic  EventTopic
	MessageData  *anypb.Any
}
