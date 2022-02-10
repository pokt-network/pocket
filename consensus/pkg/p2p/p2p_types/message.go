package p2p_types

import (
	"pocket/consensus/pkg/shared/events"
)

type NetworkMessage struct {
	Topic events.PocketEventTopic
	Data  []byte
}
