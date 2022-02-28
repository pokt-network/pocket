package types

import (
	"google.golang.org/protobuf/types/known/anypb"
)

type SourceModule string
type EventTopic string

const (
	// Core
	Consensus   SourceModule = "consensus"
	P2P         SourceModule = "p2p"
	Persistence SourceModule = "persistence"
	Utility     SourceModule = "utility"

	// Auxiliary
	Test  SourceModule = "test"
	Debug SourceModule = "debug"
)

const (
	// Consensus
	ConsensusMessage EventTopic = "CONSENSUS_MESSAGE"
)

type Event struct {
	SourceModule SourceModule
	PocketTopic  EventTopic
	MessageData  *anypb.Any
}
