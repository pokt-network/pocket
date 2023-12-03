package rpc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

type testRelayHandler func(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error)

func TestRPCServer_PostV1Client(t *testing.T) {
	testCases := []struct {
		name             string
		relay            RelayRequest
		handler          testRelayHandler
		expectedResponse string
	}{
		{
			name: "JSONRPC payload is processed correctly",
			relay: RelayRequest{
				Payload: JSONRPCPayload{
					Jsonrpc: "2.0",
					Method:  "eth_blockNumber",
					Id:      &JsonRpcId{Id: []byte("1")},
				},
			},
			handler: func(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error) {
				payload := relay.GetJsonRpcPayload()
				require.EqualValues(t, payload.Id, []byte("1"))
				require.Equal(t, payload.JsonRpc, "2.0")
				require.Equal(t, payload.Method, "eth_blockNumber")
				return &coreTypes.RelayResponse{Payload: "JSONRPC Relay Response"}, nil
			},
			expectedResponse: `{"payload":"JSONRPC Relay Response","servicer_signature":""}`,
		},
		{
			name: "REST payload is processed correctly",
			relay: RelayRequest{
				Payload: RESTPayload(`{"field1":"value1"}`),
			},
			handler: func(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error) {
				payload := relay.GetRestPayload()
				require.Equal(t, &coreTypes.RESTPayload{Contents: []byte(`{"field1":"value1"}`)}, payload)
				return &coreTypes.RelayResponse{Payload: "REST Relay Response"}, nil
			},
			expectedResponse: `{"payload":"REST Relay Response","servicer_signature":""}`,
		},
		{
			name: "Invalid payload is rejected",
			relay: RelayRequest{
				Payload: "foo",
			},
			handler:          func(_ *coreTypes.Relay) (*coreTypes.RelayResponse, error) { return nil, nil },
			expectedResponse: "bad request",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bz, err := json.Marshal(testCase.relay)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/v1/relay", bytes.NewReader(bz))
			req.Header.Add("Content-Type", "application/json")

			responseRecorder := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, responseRecorder)

			mockBus := mockBus(t, testCase.handler)
			rpcServer := NewRPCServer(mockBus)

			err = rpcServer.PostV1ClientRelay(ctx)
			require.NoError(t, err)

			resp := responseRecorder.Result()
			defer resp.Body.Close()
			responseBody, err := io.ReadAll(resp.Body)

			require.EqualValues(t, testCase.expectedResponse, strings.TrimRight(string(responseBody), "\n"))
		})
	}
}

// Create a mockBus with mock implementations of the utility module
func mockBus(t *testing.T, handler testRelayHandler) *mockModules.MockBus {
	ctrl := gomock.NewController(t)
	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().GetUtilityModule().Return(baseUtilityMock(ctrl, handler)).AnyTimes()
	return busMock
}

// Creates utility and servicer modules mocks with mock implementations of some basic functionality
func baseUtilityMock(ctrl *gomock.Controller, handler testRelayHandler) *mockModules.MockUtilityModule {
	servicerMock := mockModules.NewMockServicerModule(ctrl)
	servicerMock.EXPECT().
		HandleRelay(gomock.Any()).
		DoAndReturn(
			func(relay *coreTypes.Relay) (*coreTypes.RelayResponse, error) {
				return handler(relay)
			}).AnyTimes()

	utilityMock := mockModules.NewMockUtilityModule(ctrl)
	utilityMock.EXPECT().GetServicerModule().Return(servicerMock, nil).AnyTimes()
	return utilityMock
}
