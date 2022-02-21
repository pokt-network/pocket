package types

import (
	"pocket/shared/types"
)

type P2PMessage struct {
	Topic types.EventTopic
	Data  []byte
}
