package client

import (
	light_client_types "github.com/pokt-network/pocket/ibc/client/light_clients/types"
	"github.com/pokt-network/pocket/ibc/client/types"
	"github.com/pokt-network/pocket/ibc/path"
	"github.com/pokt-network/pocket/shared/codec"
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

// GetHostConsensusState returns the ConsensusState at the given height for the
// host chain, the Pocket network. It then serialises this and packs it into a
// ConsensusState object for use in a WASM client
func (c *clientManager) GetHostConsensusState(height modules.Height) (modules.ConsensusState, error) {
	blockStore := c.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockStore.GetBlock(height.GetRevisionHeight())
	if err != nil {
		return nil, err
	}
	pocketConsState := &light_client_types.PocketConsensusState{
		Timestamp:       block.BlockHeader.Timestamp,
		StateHash:       block.BlockHeader.StateHash,
		StateTreeHashes: block.BlockHeader.StateTreeHashes,
		NextValSetHash:  block.BlockHeader.NextValSetHash,
	}
	consBz, err := codec.GetCodec().Marshal(pocketConsState)
	if err != nil {
		return nil, err
	}
	return types.NewConsensusState(consBz, uint64(pocketConsState.Timestamp.AsTime().UnixNano())), nil
}
