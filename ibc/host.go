// TECHDEBT(#854): The host should be a submodule with access to the bus, this allows for the
// host to access the persistence module and thus the tree store in order to create local
// copies of the IBC state tree with the GetProvableStore() function. Bus access also will
// allow for the host to send any local changes to these stores as IBCMessage types through
// the P2P module's Broadcast() function to allow them to propagate through the network's mempool.
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

// GetProvableStore returns a copy of the IBC state tree where all operations observed by
// this specific ibc light client were applied to its ephemeral (in memory) state and have not
// yet been included in the next block. The aggregation of all light client-provable stores
// propagated throughout the network are happen validated by the block proposer when reaping
// the mempool, and lead to a valid on-chain state transition when consensus is reached.
// TODO(#854): Implement this
func (h *host) GetProvableStore(prefix coreTypes.CommitmentPrefix) (modules.ProvableStore, error) {
	return nil, nil
}
