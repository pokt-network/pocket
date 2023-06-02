package ibc

import (
	"time"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.IBCHost = (*Host)(nil)

type Host struct {
	logger *modules.Logger

	stores modules.StoreManager
}

// GetStoreManager returns the store manager for the host
func (h *Host) GetStoreManager() modules.StoreManager {
	return h.stores
}

// GetTimestamp returns the current unix timestamp
func (h *Host) GetTimestamp() uint64 {
	return uint64(time.Now().Unix())
}
