package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRelay_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		relay    Relay
		expected error
	}{
		{
			name: "valid Relay: JSONRPC",
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
			expected: errInvalidRelayPayload,
		},
		{
			name: "invalid Relay: invalid JSONRPC Payload",
			relay: Relay{
				RelayPayload: &Relay_JsonRpcPayload{
					JsonRpcPayload: &JSONRPCPayload{JsonRpc: "foo"},
				},
			},
			expected: errInvalidJSONRPC,
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
			require.ErrorIs(t, err, testCase.expected)
		})
	}
}

func TestRelay_ValidateJsonRpc(t *testing.T) {
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
			expected: errInvalidJSONRPC,
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
			require.ErrorIs(t, err, testCase.expected)
		})
	}
}

func TestRelay_ValidateREST(t *testing.T) {
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
			require.ErrorIs(t, err, testCase.expected)
		})
	}
}
