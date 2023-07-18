package path

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// ICS02
// The following paths are the keys to the store as defined in
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-002-client-semantics#path-space
////////////////////////////////////////////////////////////////////////////////

// FullClientStateKey takes a client identifier and returns a Key under which to store a
// particular client state.
func FullClientStateKey(clientID string) []byte {
	return fullClientKey(clientID, KeyClientState)
}

// ClientStatePath takes a client identifier and returns a Path string where it can be accessed
// within the client store
func ClientStatePath(clientID string) string {
	return clientPath(clientID, KeyClientState)
}

// consensusStatePath returns the suffix store key for the consensus state at a
// particular height stored in a client prefixed store.
func consensusStatePath(height string) string {
	return fmt.Sprintf("%s/%s", KeyConsensusStatePrefix, height)
}

// ConsensusStateKey returns the store key for the consensus state of a particular client
// in a prefixed client store
func ConsensusStateKey(height string) []byte {
	return []byte(consensusStatePath(height))
}

// fullConsensusStatePath takes a client identifier and returns a Path under which to
// store the consensus state of a client.
func fullConsensusStatePath(clientID, height string) string {
	return fullClientPath(clientID, consensusStatePath(height))
}

// FullConsensusStateKey returns the store key for the consensus state of a particular client.
func FullConsensusStateKey(clientID, height string) []byte {
	return []byte(fullConsensusStatePath(clientID, height))
}

// ConsensusStatePath takes a client identifier and height and returns the Path where the consensus
// state can be accessed in the client store
func ConsensusStatePath(clientID, height string) string {
	return clientPath(clientID, consensusStatePath(height))
}
