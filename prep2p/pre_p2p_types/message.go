package pre_p2p_types

import (
	"pocket/shared/events"
)

type NetworkMessage struct {
	Topic events.PocketEventTopic
	Data  []byte
}
