package ibc

import (
	"time"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.IBCHost = &host{}

type host struct {
	logger *modules.Logger
}

// GetTimestamp returns the current unix timestamp
func (h *host) GetTimestamp() uint64 {
	return uint64(time.Now().Unix())
}
