package types

import (
	"math/rand"
	"time"

	shared "github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

func GenerateRandomHash() int64 {
	rand.Seed(time.Now().UnixNano())
	return int64(rand.Int())
}

func NewP2PMessage(nonce int32, level int32, src, dest string, event *shared.PocketEvent) *P2PMessage {
	return &P2PMessage{
		Metadata: &Metadata{
			// TODO(derrandz): bring back when the hash behavior is spec'd
			Hash:        GenerateRandomHash(),
			Level:       level,
			Source:      src,
			Destination: dest,
		},
		Payload: event,
	}
}

func NewPocketEvent(topic shared.PocketTopic, message *anypb.Any) *shared.PocketEvent {
	return &shared.PocketEvent{
		Topic: topic,
		Data:  message,
	}
}

func (msg *P2PMessage) MarkAsBroadcastMessage() {
	msg.Metadata.Broadcast = true
}
