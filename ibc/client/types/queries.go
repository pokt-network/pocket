package types

import (
	"github.com/pokt-network/pocket/ibc/path"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
)

// GetConsensusState returns the consensus state at the given height from a
// prefixed client store, in the format: "clients/{clientID}"
func GetConsensusState(clientStore modules.ProvableStore, height modules.Height) (modules.ConsensusState, error) {
	// Retrieve the consensus state bytes from the client store
	consStateBz, err := clientStore.Get(path.ConsensusStateKey(height.ToString()))
	if err != nil {
		return nil, err
	}

	// Unmarshal into a ConsensusState interface
	consState := new(ConsensusState)
	if err := codec.GetCodec().Unmarshal(consStateBz, consState); err != nil {
		return nil, err
	}

	return consState, nil
}

// GetClientState returns the client state from a prefixed client store,
// in the format: "clients" using the clientID provided
func GetClientState(clientStore modules.ProvableStore, identifier string) (modules.ClientState, error) {
	// Retrieve the client state bytes from the client store
	clientStateBz, err := clientStore.Get(path.FullClientStateKey(identifier))
	if err != nil {
		return nil, err
	}

	// Unmarshal into a ClientState interface
	clientState := new(ClientState)
	if err := codec.GetCodec().Unmarshal(clientStateBz, clientState); err != nil {
		return nil, err
	}

	return clientState, nil
}

// setClientState stores the client state
// clientStore must be a prefixed client store: "clients/{clientID}"
func setClientState(clientStore modules.ProvableStore, clientState *ClientState) error {
	val, err := codec.GetCodec().Marshal(clientState)
	if err != nil {
		return err
	}
	return clientStore.Set(nil, val) // key == nil ==> key == "clients/{clientID}"
}

// setConsensusState stores the consensus state at the given height.
// clientStore must be a prefixed client store: "clients/{clientID}"
func setConsensusState(clientStore modules.ProvableStore, consensusState *ConsensusState, height modules.Height) error {
	val, err := codec.GetCodec().Marshal(consensusState)
	if err != nil {
		return err
	}
	return clientStore.Set(path.ConsensusStateKey(height.ToString()), val)
}
