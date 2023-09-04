package types

import (
	"errors"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ConsensusState = &ConsensusState{}

// NewConsensusState creates a new ConsensusState instance.
func NewConsensusState(data []byte, timestamp uint64) *ConsensusState {
	return &ConsensusState{
		Data:      data,
		Timestamp: timestamp,
	}
}

// ClientType returns the Wasm client type.
func (cs *ConsensusState) ClientType() string {
	return WasmClientType
}

// ValidateBasic defines a basic validation for the wasm client consensus state.
func (cs *ConsensusState) ValidateBasic() error {
	if cs.Timestamp == 0 {
		return errors.New("timestamp must be a positive Unix time")
	}
	if len(cs.Data) == 0 {
		return errors.New("data cannot be empty")
	}
	return nil
}
