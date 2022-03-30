package types

import (
	"pocket/shared/types"
)

type NetworkMessage struct {
	Topic types.EventTopic
	Data  []byte
}
