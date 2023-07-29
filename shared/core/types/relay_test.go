package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRelay_Validate(t *testing.T) {
	testCases := []struct {
		name     string
		relay    *Relay
		expected error
	}{
		{
			name: "valid Relay: JSONRPC",
			relay: &Relay{
				RelayPayload: &Relay_JsonRpcPayload{
					JsonRpcPayload: &JSONRPCPayload{Jsonrpc: "2.0", Method: "eth_blockNumber"},
				},
			},
		},
		{
			name: "valid Relay: REST",
			relay: &Relay{
				RelayPayload: &Relay_RestPayload{
					RestPayload: &RESTPayload{Contents: `{"field1": "value1", "field2": "value2"}`},
				},
			},
		},
		{
			name:     "invalid Relay: missing payload",
			relay:    &Relay{},
			expected: errInvalidRelayPayload,
		},
		{
			name: "invalid Relay: invalid JSONRPC Payload",
			relay: &Relay{
				RelayPayload: &Relay_JsonRpcPayload{
					JsonRpcPayload: &JSONRPCPayload{Jsonrpc: "foo"},
				},
			},
			expected: errInvalidJSONRPC,
		},
		{
			name: "invalid Relay: invalid REST Payload",
			relay: &Relay{
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
		payload  *JSONRPCPayload
		expected error
	}{
		{
			name:    "valid JSONRPC",
			payload: &JSONRPCPayload{Jsonrpc: "2.0", Method: "eth_blockNumber"},
		},
		{
			name:     "invalid JSONRPC: invalid Jsonrpc field value",
			payload:  &JSONRPCPayload{Jsonrpc: "foo", Method: "eth_blockNumber"},
			expected: errInvalidJSONRPC,
		},
		{
			name:     "invalid JSONRPC: Method field not set",
			payload:  &JSONRPCPayload{Jsonrpc: "2.0"},
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
		payload  *RESTPayload
		expected error
	}{
		{
			name:    "valid REST",
			payload: &RESTPayload{Contents: `{"field1": "value1", "field2": "value2"}`},
		},
		{
			name:     "invalid REST payload: is not JSON-formatted",
			payload:  &RESTPayload{Contents: "foo"},
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

func TestRelay_MarshalJSONRPC(t *testing.T) {
	payload := JSONRPCPayload{
		Id:      &JSONRPCId{Id: []byte(`"1"`)},
		Jsonrpc: "2.0",
		Method:  "eth_blockNumber",
	}
	expected := []byte(`{"id":"1","jsonrpc":"2.0","method":"eth_blockNumber"}`)

	bz, err := json.Marshal(payload)
	require.NoError(t, err)
	require.Equal(t, bz, expected)
}
