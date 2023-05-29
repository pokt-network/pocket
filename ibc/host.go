package ibc

import (
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.IBCHost = (*Host)(nil)

type Host struct {
	logger *modules.Logger

	stores modules.StoreManager
}

func (h *Host) GetStore() modules.StoreManager {
	return h.stores
}
