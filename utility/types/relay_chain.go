package types

import "github.com/pokt-network/pocket/shared/types"

const (
	RelayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

type RelayChain string

// TODO: Consider adding a governance parameter for a list of valid relay chains
func (rc *RelayChain) Validate() types.Error {
	if rc == nil || *rc == "" {
		return types.ErrEmptyRelayChain()
	}
	rcLen := len(*rc)
	if rcLen != RelayChainLength {
		return types.ErrInvalidRelayChainLength(rcLen, RelayChainLength)
	}
	return nil
}
