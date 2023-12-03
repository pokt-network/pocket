package types

import (
	"errors"

	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ClientMessage = &Header{}

// ClientType defines that the Header is a Wasm client consensus algorithm
func (h *Header) ClientType() string {
	return WasmClientType
}

// ValidateBasic defines a basic validation for the wasm client header.
func (h *Header) ValidateBasic() error {
	if len(h.Data) == 0 {
		return errors.New("data cannot be empty")
	}

	return nil
}
