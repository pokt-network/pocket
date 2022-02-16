package pre_p2p_types

import (
	"pocket/shared/types"
)

type NetworkMessage struct {
	Topic types.PocketEventTopic
	Data  []byte
}
