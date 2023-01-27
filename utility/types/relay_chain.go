package types

type RelayChain string

const (
	RelayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

// TODO: Consider adding a governance parameter for a list of valid relay chains
func (rc *RelayChain) Validate() Error {
	if rc == nil || *rc == "" {
		return ErrEmptyRelayChain()
	}
	rcLen := len(*rc)
	if rcLen != RelayChainLength {
		return ErrInvalidRelayChainLength(rcLen, RelayChainLength)
	}
	return nil
}
