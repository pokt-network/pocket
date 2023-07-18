package types

import (
	"errors"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ClientMessage = (*Misbehaviour)(nil)

// ClientType is Wasm light client
func (m *Misbehaviour) ClientType() string {
	return WasmClientType
}

// ValidateBasic implements Misbehaviour interface
func (m *Misbehaviour) ValidateBasic() error {
	if len(m.Data) == 0 {
		return errors.New("data cannot be empty")
	}
	return nil
}
