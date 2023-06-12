package service

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

const (
	testAppsTokensMultiplier = int(2)
	testCurrentHeight        = int64(9)
)

var (
	testServicer1 = &coreTypes.Actor{
		ActorType: coreTypes.ActorType_ACTOR_TYPE_SERVICER,
		Address:   "a3d9ea9d9ad9c58bb96ec41340f83cb2cabb6496",
		PublicKey: "a6cd0a304c38d76271f74dd3c90325144425d904ef1b9a6fbab9b201d75a998b",
		Chains:    []string{"0021"},
	}

	testApp1 = &coreTypes.Actor{
		Address:      "98a792b7aca673620132ef01f50e62caa58eca83",
		PublicKey:    "b5cd0a304c38d76271f74dd3c90325144425d904ef1b9a6fbab9b201d86b009c",
		StakedAmount: "1000",
	}

	testServiceConfig1 = &configs.ServiceConfig{
		Url:         "http://chain-url.pokt.network",
		TimeoutMsec: 1234,
		BasicAuth: &configs.BasicAuth{
			UserName: "user1",
			Password: "password1",
		},
	}
)

func TestAdmitRelay(t *testing.T) {
	const (
		currentSessionNumber      = 2
		testSessionStartingHeight = 8
	)

	testCases := []struct {
		name              string
		usedSessionTokens int64
		relay             *coreTypes.Relay
		expected          error
	}{
		{
			name:  "valid relay is admitted",
			relay: testRelay(),
		},
		{
			name:     "Relay with empty Meta is rejected",
			relay:    &coreTypes.Relay{},
			expected: errValidateRelayMeta,
		},
		{
			name:     "Relay with unspecified chain is rejected",
			relay:    &coreTypes.Relay{Meta: &coreTypes.RelayMeta{}},
			expected: errValidateRelayMeta,
		},
		{
			name:     "Relay for unsupported chain is rejected",
			relay:    testRelay(testRelayChain("foo")),
			expected: errValidateRelayMeta,
		},
		{
			name:     "Relay with height set in a past session is rejected",
			relay:    testRelay(testRelayHeight(5)),
			expected: errValidateBlockHeight,
		},
		{
			name:     "Relay with height set in a future session is rejected",
			relay:    testRelay(testRelayHeight(9999)),
			expected: errValidateBlockHeight,
		},
		{
			name:     "Relay not matching the servicer in this session is rejected",
			relay:    testRelay(testRelayServicer("bar")),
			expected: errValidateServicer,
		},
		{
			name:              "Relay for app out of quota is rejected",
			relay:             testRelay(),
			usedSessionTokens: 999999,
			expected:          errValidateApplication,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			config := testServicerConfig()
			session := testSession(
				sessionNumber(currentSessionNumber),
				sessionBlocks(4),
				sessionHeight(testSessionStartingHeight),
				sessionServicers(testServicer1),
			)
			mockBus := mockBus(t, &config, uint64(testCurrentHeight), session, testCase.usedSessionTokens)

			servicerMod, err := CreateServicer(mockBus)
			require.NoError(t, err)

			servicer, ok := servicerMod.(*servicer)
			require.True(t, ok)

			err = servicer.admitRelay(testCase.relay)
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("Expected error %v got: %v", testCase.expected, err)
			}
		})
	}
}

func TestExecuteRelay(t *testing.T) {
	testCases := []struct {
		name        string
		relay       *coreTypes.Relay
		expectedErr error
	}{
		{
			name:        "relay is rejected if chain is not specified in the config",
			relay:       testRelay(testRelayChain("foo")),
			expectedErr: errValidateRelayMeta,
		},
		{
			name:  "Relay for accepted chain is executed",
			relay: testRelay(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"0x1234"}`)
			}))
			defer ts.Close()

			config := testServicerConfig()
			for ch := range config.Services {
				config.Services[ch].Url = ts.URL
			}

			servicer := &servicer{config: &config}
			_, err := servicer.executeRelay(testCase.relay)
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("Expected error %v got: %v", testCase.expectedErr, err)
			}
			// INCOMPLETE: verify HTTP request properties: payload/headers/user-agent/etc.
		})
	}
}

type relayEditor func(*coreTypes.Relay)

func testRelayServicer(publicKey string) relayEditor {
	return func(relay *coreTypes.Relay) {
		relay.Meta.ServicerPublicKey = publicKey
	}
}

func testRelayChain(chain string) relayEditor {
	return func(relay *coreTypes.Relay) {
		relay.Meta.RelayChain = &coreTypes.Identifiable{Id: chain}
	}
}

func testRelayHeight(height int64) relayEditor {
	return func(relay *coreTypes.Relay) {
		relay.Meta.BlockHeight = height
	}
}

func testRelay(editors ...relayEditor) *coreTypes.Relay {
	relay := &coreTypes.Relay{
		Meta: &coreTypes.RelayMeta{
			ServicerPublicKey:  testServicer1.PublicKey,
			ApplicationAddress: testApp1.Address,
			BlockHeight:        testCurrentHeight,
			RelayChain: &coreTypes.Identifiable{
				Id: "0021",
			},
			GeoZone: &coreTypes.Identifiable{
				Id: "geozone",
			},
		},
		RelayPayload: &coreTypes.Relay_JsonRpcPayload{
			&coreTypes.JsonRpcPayload{
				Method: "POST",
				Data:   []byte(`{"id": 1, "jsonrpc": "2.0", method: "eth_blockNumber"}`),
			},
		},
	}

	for _, editor := range editors {
		editor(relay)
	}

	return relay
}

func testServicerConfig() configs.ServicerConfig {
	return configs.ServicerConfig{
		PublicKey: testServicer1.PublicKey,
		Address:   testServicer1.Address,
		Services: map[string]*configs.ServiceConfig{
			"0021": testServiceConfig1,
		},
	}
}

type sessionModifier func(*coreTypes.Session)

func sessionNumber(number int64) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.SessionNumber = number
	}
}

func sessionBlocks(blocksPerSession int64) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.NumSessionBlocks = blocksPerSession
	}
}

func sessionHeight(height int64) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.SessionHeight = height
	}
}

func sessionServicers(servicers ...*coreTypes.Actor) func(*coreTypes.Session) {
	return func(session *coreTypes.Session) {
		session.Servicers = servicers
	}
}

func testSession(editors ...sessionModifier) *coreTypes.Session {
	session := coreTypes.Session{
		Id:          "session-1",
		Application: testApp1,
	}
	for _, editor := range editors {
		editor(&session)
	}
	return &session
}

// Create a mockBus with mock implementations of consensus and utility modules
func mockBus(t *testing.T, cfg *configs.ServicerConfig, height uint64, session *coreTypes.Session, usedSessionTokens int64) *mockModules.MockBus {
	ctrl := gomock.NewController(t)
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(&configs.Config{Utility: &configs.UtilityConfig{ServicerConfig: cfg}}).AnyTimes()

	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(height).AnyTimes()

	persistenceReadContextMock := mockModules.NewMockPersistenceReadContext(ctrl)
	persistenceReadContextMock.EXPECT().Release().AnyTimes()
	persistenceReadContextMock.EXPECT().GetServicer(gomock.Any(), gomock.Any()).Return(testServicer1, nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetIntParam(typesUtil.AppSessionTokensMultiplierParamName, session.SessionHeight).
		Return(testAppsTokensMultiplier, nil).AnyTimes()

	persistenceLocalContextMock := mockModules.NewMockPersistenceLocalContext(ctrl)
	persistenceLocalContextMock.EXPECT().StoreServiceRelay(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	persistenceLocalContextMock.EXPECT().GetSessionTokensUsed(gomock.Any()).Return(big.NewInt(usedSessionTokens), nil).AnyTimes()
	persistenceLocalContextMock.EXPECT().Release().Return(nil).AnyTimes()

	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()
	persistenceMock.EXPECT().NewLocalContext().Return(persistenceLocalContextMock, nil).AnyTimes()

	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	busMock.EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetUtilityModule().Return(baseUtilityMock(ctrl, session)).AnyTimes()
	busMock.EXPECT().RegisterModule(gomock.Any()).DoAndReturn(func(m modules.Module) {
		m.SetBus(busMock)
	}).AnyTimes()

	return busMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(ctrl *gomock.Controller, session *coreTypes.Session) *mockModules.MockUtilityModule {
	utilityMock := mockModules.NewMockUtilityModule(ctrl)
	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	utilityMock.EXPECT().
		GetSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(
			// mimicking the behavior of the utility module's GetSession method
			// IMPROVE: verify values passed to GetSession
			// IMPROVE: use editor functions to allow the test case to modify the session.
			func(appAddr string, height int64, relayChain, geoZone string) (*coreTypes.Session, error) {
				return session, nil
			}).
		MaxTimes(1)
	utilityMock.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()

	return utilityMock
}
