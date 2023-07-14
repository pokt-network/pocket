package client

import (
	"github.com/pokt-network/pocket/ibc/path"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
)

// GetConsensusState returns the ConsensusState at the given height for the
// stored client with the given identifier
func (c *clientManager) GetConsensusState(
	identifier string, height modules.Height,
) (modules.ConsensusState, error) {
	// Retrieve the client store
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(path.KeyClientStorePrefix)
	if err != nil {
		return nil, err
	}

	// Retrieve the consensus state bytes from the client store
	consStateBz, err := clientStore.Get(path.FullConsensusStateKey(identifier, height.ToString()))
	if err != nil {
		return nil, err
	}

	// Unmarshal into a ConsensusState interface
	var consState modules.ConsensusState
	if err := codec.GetInterfaceRegistry().UnmarshalInterface(consStateBz, &consState); err != nil {
		return nil, err
	}

	return consState, nil
}

// GetClientState returns the ClientState for the stored client with the given identifier
func (c *clientManager) GetClientState(identifier string) (modules.ClientState, error) {
	// Retrieve the client store
	clientStore, err := c.GetBus().GetIBCHost().GetProvableStore(path.KeyClientStorePrefix)
	if err != nil {
		return nil, err
	}

	// Retrieve the client state bytes from the client store
	clientStateBz, err := clientStore.Get(path.FullClientStateKey(identifier))
	if err != nil {
		return nil, err
	}

	// Unmarshal into a ClientState interface
	var clientState modules.ClientState
	if err := codec.GetInterfaceRegistry().UnmarshalInterface(clientStateBz, &clientState); err != nil {
		return nil, err
	}

	return clientState, nil
}
