package types

const (
	// DISCUSS: Should this be a governance parameter or moved to a shared file?
	relayChainLength = 4 // pre-determined length that strikes a balance between combination possibilities & storage
)

type relayChain string

// No need for a relayChain interface abstraction for the time being
var _ Validatable = relayChain("")

// TODO: Consider adding a governance parameter for a list of valid relay chains
func (rc relayChain) ValidateBasic() Error {
	if rc == "" {
		return ErrEmptyRelayChain()
	}
	rcLen := len(rc)
	if rcLen != relayChainLength {
		return ErrInvalidRelayChainLength(rcLen, relayChainLength)
	}
	return nil
}
