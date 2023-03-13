package types

import (
	"github.com/pokt-network/pocket/shared/pokterrors"
)

const (
	// DISCUSS: Should this be a governance parameter or moved to a shared file?
	relayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

type relayChain string

// TODO: Consider adding a governance parameter for a list of valid relay chains
// ValidateBasic validates the relay chain follows a pre-determined format
func (rc relayChain) ValidateBasic() pokterrors.Error {
	if rc == "" {
		return pokterrors.UtilityErrEmptyRelayChain()
	}
	rcLen := len(rc)
	if rcLen != relayChainLength {
		return pokterrors.UtilityErrInvalidRelayChainLength(rcLen, relayChainLength)
	}
	return nil
}
