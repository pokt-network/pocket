package types

import coreTypes "github.com/pokt-network/pocket/shared/core/types"

const (
	// DISCUSS: Should this be a governance parameter or moved to a shared file?
	relayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

type relayChain string

// TODO: Consider adding a governance parameter for a list of valid relay chains
// ValidateBasic validates the relay chain follows a pre-determined format
func (rc relayChain) ValidateBasic() coreTypes.Error {
	if rc == "" {
		return coreTypes.ErrEmptyRelayChain()
	}
	rcLen := len(rc)
	if rcLen != relayChainLength {
		return coreTypes.ErrInvalidRelayChainLength(rcLen, relayChainLength)
	}
	return nil
}
