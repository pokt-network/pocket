package ibc

import (
	"time"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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

// GetProvableStore returns
func (h *host) GetProvableStore(prefix coreTypes.CommitmentPrefix) (modules.ProvableStore, error) {
	return nil, nil
}
