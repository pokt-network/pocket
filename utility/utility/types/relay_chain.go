package types

const (
	RelayChainLength = 2
)

type RelayChain string

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
