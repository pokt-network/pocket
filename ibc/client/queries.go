package client

import (
	"github.com/pokt-network/pocket/ibc/client/types"
	"github.com/pokt-network/pocket/ibc/path"
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// GetConsensusState returns the ConsensusState at the given height for the
// stored client with the given identifier
func (c *clientManager) GetConsensusState(
	identifier string, height modules.Height,
) (modules.ConsensusState, error) {
	// Retrieve the clientId prefixed client store
	prefixed := path.ApplyPrefix(core_types.CommitmentPrefix(path.KeyClientStorePrefix), identifier)
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(string(prefixed))
	if err != nil {
		return nil, err
	}

	return types.GetConsensusState(clientStore, height)
}

// GetClientState returns the ClientState for the stored client with the given identifier
func (c *clientManager) GetClientState(identifier string) (modules.ClientState, error) {
	// Retrieve the client store
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(path.KeyClientStorePrefix)
	if err != nil {
		return nil, err
	}

	return types.GetClientState(clientStore, identifier)
}
