// TECHDEBT(#854): The host should be a submodule with access to the bus, this allows for the
// host to access the persistence module and thus the tree store in order to create local
// copies of the IBC state tree with the GetProvableStore() function. Bus access also will
// allow for the host to send any local changes to these stores as IbcMessage types through
// the P2P module's Broadcast() function to allow them to be included in the next block.
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

// GetProvableStore returns a copy of the IBC state tree where all operations happen
// locally (in memory) and are not persisted to the database. All changes are instead
// broadcasted to the network for inclusion in the next block.
// TODO(#854): Implement this
func (h *host) GetProvableStore(prefix coreTypes.CommitmentPrefix) (modules.ProvableStore, error) {
	return nil, nil
}
