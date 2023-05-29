package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	errInvalidRelayInvalidPayload  = errors.New("invalid relay payload")
	errInvalidJSONRPCInvalidRPC    = errors.New("invalid value for JSONRPC field")
	errInvalidJSONRPCMissingMethod = errors.New("Method field not set")
	errInvalidRESTPayload          = errors.New("invalid REST payload")
)

// INCOMPLETE: perform any possible metadata validation
// Validate performs validation on the relay payload
func (r Relay) Validate() error {
	if jsonRpcPayload := r.GetJsonRpcPayload(); jsonRpcPayload != nil {
		return jsonRpcPayload.Validate()
	}

	if jsonPayload := r.GetRestPayload(); jsonPayload != nil {
		return jsonPayload.Validate()
	}

	return fmt.Errorf("%w: %v", errInvalidRelayInvalidPayload, r)
}

// Validate performs validation on JSONRPC payload. More specifically, it verifies that:
//  1. The JSONRPC field is set to "2.0" as per the JSONRPC spec requirement, and
//  2. The Method field is not empty
func (p JSONRPCPayload) Validate() error {
	if p.JsonRpc != "2.0" {
		return fmt.Errorf("%w: %s", errInvalidJSONRPCInvalidRPC, p.JsonRpc)
	}

	// DISCUSS: do we need/want chain-specific validation? Potential for reusing the existing logic of Portal V2/pocket-go
	if p.Method == "" {
		return errInvalidJSONRPCMissingMethod
	}
	return nil
}

// Validate verifies that the payload is valid REST, i.e. valid JSON
func (p RESTPayload) Validate() error {
	var parsed json.RawMessage
	err := json.Unmarshal([]byte(p.Contents), &parsed)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", errInvalidRESTPayload, p.Contents, err)
	}
	return nil
}
