package types

import "github.com/pokt-network/pocket/shared/types"

const (
	RelayChainLength = 4
)

type RelayChain string

func (rc *RelayChain) Validate() types.Error {
	// TODO: Consider validating against a static list (gov parameter?) of supported relay chains
	if rc == nil || *rc == "" {
		return types.ErrEmptyRelayChain()
	}
	rcLen := len(*rc)
	if rcLen != RelayChainLength {
		return types.ErrInvalidRelayChainLength(rcLen, RelayChainLength)
	}
	return nil
}
