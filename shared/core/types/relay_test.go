package types

import (
	"errors"
	"testing"
)

func TestRelayValidate(t *testing.T) {
	testCases := []struct {
		name     string
		relay    Relay
		expected error
	}{
		{
			name: "valid Relay: JSONRPC",
			// IMPROVE: use a factory function to build test relays
			relay: Relay{
				RelayPayload: &Relay_JsonRpcPayload{
					JsonRpcPayload: &JSONRPCPayload{JsonRpc: "2.0", Method: "eth_blockNumber"},
				},
			},
		},
		{
			name: "valid Relay: REST",
			relay: Relay{
				RelayPayload: &Relay_RestPayload{
					RestPayload: &RESTPayload{Contents: `{"field1": "value1", "field2": "value2"}`},
				},
			},
		},
		{
			name:     "invalid Relay: missing payload",
			expected: errInvalidRelayInvalidPayload,
		},
		{
			name: "invalid Relay: invalid JSONRPC Payload",
			relay: Relay{
				RelayPayload: &Relay_JsonRpcPayload{
					JsonRpcPayload: &JSONRPCPayload{JsonRpc: "foo"},
				},
			},
			expected: errInvalidJSONRPCInvalidRPC,
		},
		{
			name: "invalid Relay: invalid REST Payload",
			relay: Relay{
				RelayPayload: &Relay_RestPayload{
					RestPayload: &RESTPayload{Contents: "foo"},
				},
			},
			expected: errInvalidRESTPayload,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.relay.Validate()
			switch {
			case err == nil && testCase.expected != nil:
				t.Fatalf("Expected error %v, got: nil", testCase.expected)
			case err != nil && testCase.expected == nil:
				t.Fatalf("Unexpected error: %v", err)
			case testCase.expected != nil && !errors.Is(err, testCase.expected):
				t.Fatalf("Expected error %v got: %v", testCase.expected, err)
			}
		})
	}
}

func TestJsonRpcValidate(t *testing.T) {
	testCases := []struct {
		name     string
		payload  JSONRPCPayload
		expected error
	}{
		{
			name:    "valid JSONRPC",
			payload: JSONRPCPayload{JsonRpc: "2.0", Method: "eth_blockNumber"},
		},
		{
			name:     "invalid JSONRPC: invalid JsonRpc field value",
			payload:  JSONRPCPayload{JsonRpc: "foo", Method: "eth_blockNumber"},
			expected: errInvalidJSONRPCInvalidRPC,
		},
		{
			name:     "invalid JSONRPC: Method field not set",
			payload:  JSONRPCPayload{JsonRpc: "2.0"},
			expected: errInvalidJSONRPCMissingMethod,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.payload.Validate()
			switch {
			case err == nil && testCase.expected != nil:
				t.Fatalf("Expected error %v, got: nil", testCase.expected)
			case err != nil && testCase.expected == nil:
				t.Fatalf("Unexpected error: %v", err)
			case testCase.expected != nil && !errors.Is(err, testCase.expected):
				t.Fatalf("Expected error %v got: %v", testCase.expected, err)
			}
		})
	}
}

func TestRESTValidate(t *testing.T) {
	testCases := []struct {
		name     string
		payload  RESTPayload
		expected error
	}{
		{
			name:    "valid REST",
			payload: RESTPayload{Contents: `{"field1": "value1", "field2": "value2"}`},
		},
		{
			name:     "invalid REST payload: is not JSON-formatted",
			payload:  RESTPayload{Contents: "foo"},
			expected: errInvalidRESTPayload,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.payload.Validate()
			switch {
			case err == nil && testCase.expected != nil:
				t.Fatalf("Expected error %v, got: nil", testCase.expected)
			case err != nil && testCase.expected == nil:
				t.Fatalf("Unexpected error: %v", err)
			case testCase.expected != nil && !errors.Is(err, testCase.expected):
				t.Fatalf("Expected error %v got: %v", testCase.expected, err)
			}
		})
	}
}
