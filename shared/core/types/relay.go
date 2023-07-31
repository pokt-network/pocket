package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

const jsonRpcVersion = "2.0"

var (
	// TODO: Once the proto structures are finalized, add an exhaustive list of errors and tests
	errInvalidRelayPayload         = errors.New("invalid relay payload")
	errInvalidJSONRPC              = errors.New("invalid value for JSONRPC field")
	errInvalidJSONRPCMissingMethod = errors.New("Method field not set")
	errInvalidRESTPayload          = errors.New("invalid REST payload")
)

// IMPROVE: use a factory function to build test relays
// INCOMPLETE: perform any possible metadata validation
// Validate performs validation on the relay payload
func (r *Relay) Validate() error {
	switch payload := r.RelayPayload.(type) {
	case *Relay_JsonRpcPayload:
		return payload.JsonRpcPayload.Validate()
	case *Relay_RestPayload:
		return payload.RestPayload.Validate()
	default:
		return fmt.Errorf("%w: %v", errInvalidRelayPayload, r)
	}
}

// Validate performs validation on JSONRPC payload. More specifically, it verifies that:
//  1. The JSONRPC field is set to "2.0" as per the JSONRPC spec requirement, and
//  2. The Method field is not empty
func (p *JSONRPCPayload) Validate() error {
	if p.Jsonrpc != jsonRpcVersion {
		return fmt.Errorf("%w: %s", errInvalidJSONRPC, p.Jsonrpc)
	}

	// DISCUSS: do we need/want chain-specific validation? Potential for reusing the existing logic of Portal V2/pocket-go
	//	Potential items to consider when validating: number of RelayChains to stake, permissioned RelayChains, other types of services, e.g. <Type>.<ID>.<Config>
	if p.Method == "" {
		return errInvalidJSONRPCMissingMethod
	}
	return nil
}

// Validate verifies that the payload is valid REST, i.e. valid JSON
func (p *RESTPayload) Validate() error {
	var parsed json.RawMessage
	if err := json.Unmarshal([]byte(p.Contents), &parsed); err != nil {
		return fmt.Errorf("%w: %s: %s", errInvalidRESTPayload, p.Contents, err.Error())
	}
	return nil
}

// MarshalJSON is a custom marshaller for JSONRPCId type to return the byte array as-is.
//
//	This is to ensure the specified ID gets sent correctly when serializing the relay that contains the ID.
func (i *JSONRPCId) MarshalJSON() ([]byte, error) {
	return i.Id, nil
}
