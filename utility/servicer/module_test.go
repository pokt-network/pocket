package servicer

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

const (
	testAppsTokensMultiplier = int(2)
	testCurrentHeight        = int64(9)
)

// INCOMPLETE(#833) add e2e testing on servicer's features

var (
	// Initialized in TestMain
	testServicer1           *coreTypes.Actor
	testServicer1PrivateKey crypto.PrivateKey

	// Initialized in TestMain
	testApp1 *coreTypes.Actor

	// Initialized in TestMain
	testServiceConfig1 *configs.ServiceConfig
)

// testPublicKey is a helper that returns a public key and its corresponding address
func testPublicKey() (publicKey, address string) {
	pk, err := crypto.GeneratePublicKey()
	if err != nil {
		log.Fatalf("Error creating public key: %s", err)
	}

	return pk.String(), pk.Address().String()
}

// TestMain initialized the test fixtures for all the unit tests in the servicer package
func TestMain(m *testing.M) {
	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("Error generating private key: %s", err)
	}

	testServicer1PrivateKey = privateKey
	testServicer1 = &coreTypes.Actor{
		ActorType:    coreTypes.ActorType_ACTOR_TYPE_SERVICER,
		Address:      privateKey.Address().String(),
		PublicKey:    privateKey.PublicKey().String(),
		Chains:       []string{"POKT-UnitTestNet"},
		StakedAmount: "1000",
	}

	appPublicKey, appAddr := testPublicKey()
	testApp1 = &coreTypes.Actor{
		ActorType:    coreTypes.ActorType_ACTOR_TYPE_APP,
		Address:      appAddr,
		PublicKey:    appPublicKey,
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

	os.Exit(m.Run())
}

func TestRelay_Admit(t *testing.T) {
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
			name:     "Relay for unsupported service is rejected",
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
			expected:          errShouldMineRelay,
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
			mockBus := mockBus(t, config, uint64(testCurrentHeight), session, testCase.usedSessionTokens)

			servicerMod, err := CreateServicer(mockBus)
			require.NoError(t, err)

			servicer, ok := servicerMod.(*servicer)
			require.True(t, ok)

			err = servicer.admitRelay(testCase.relay)
			require.ErrorIs(t, err, testCase.expected)
		})
	}
}

func TestRelay_Execute(t *testing.T) {
	testCases := []struct {
		name        string
		relay       *coreTypes.Relay
		expectedErr error
	}{
		{
			name:        "relay is rejected if service is not specified in the config",
			relay:       testRelay(testRelayChain("foo")),
			expectedErr: errValidateRelayMeta,
		},
		{
			name:  "Relay for accepted service is executed",
			relay: testRelay(),
		},
		{
			name:  "JSONRPC Relay is executed",
			relay: testRelay(testEthGoerliRelay()),
		},
		{
			name:  "REST Relay is executed",
			relay: testRelay(testRESTRelay()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"0x1234"}`)
			}))
			defer ts.Close()

			config := testServicerConfig()
			for svc := range config.Services {
				config.Services[svc].Url = ts.URL
			}

			servicer := &servicer{config: config}
			_, err := servicer.executeRelay(testCase.relay)
			require.ErrorIs(t, err, testCase.expectedErr)
			// INCOMPLETE(@adshmh): verify HTTP request properties: payload/headers/user-agent/etc.
		})
	}
}

func TestRelay_Sign(t *testing.T) {
	testCases := []struct {
		name       string
		privateKey string
		expected   []byte
		expectErr  bool
	}{
		{
			name:      "Create fails if private key is missing from config",
			expectErr: true,
		},
		{
			name:       "Message is signed using correct private key",
			privateKey: testServicer1PrivateKey.String(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			config := testServicerConfig(withPrivateKey(testCase.privateKey))
			mockBus := mockBus(t, config, 0, &coreTypes.Session{}, 0)

			servicerMod, err := CreateServicer(mockBus)
			if testCase.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			servicer, ok := servicerMod.(*servicer)
			require.True(t, ok)

			message := []byte("message")
			signature, err := servicer.sign(message)
			require.NoError(t, err)

			isSignatureValid := testServicer1PrivateKey.PublicKey().Verify(message, signature)
			require.True(t, isSignatureValid)
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

func testEthGoerliRelay() relayEditor {
	return func(relay *coreTypes.Relay) {
		relay.Meta.RelayChain.Id = "ETH-Goerli"
		relay.RelayPayload = &coreTypes.Relay_JsonRpcPayload{
			JsonRpcPayload: &coreTypes.JSONRPCPayload{
				Id:      []byte("1"),
				JsonRpc: "2.0",
				Method:  "eth_blockNumber",
			},
		}
	}
}

func testRESTRelay() relayEditor {
	return func(relay *coreTypes.Relay) {
		relay.Meta.RelayChain.Id = "RESTful-service"
		relay.RelayPayload = &coreTypes.Relay_RestPayload{
			RestPayload: &coreTypes.RESTPayload{
				Contents: []byte(`{"field1": "value1"}`),
			},
		}
	}
}

func testRelay(editors ...relayEditor) *coreTypes.Relay {
	relay := &coreTypes.Relay{
		Meta: &coreTypes.RelayMeta{
			ServicerPublicKey:  testServicer1.PublicKey,
			ApplicationAddress: testApp1.Address,
			BlockHeight:        testCurrentHeight,
			RelayChain: &coreTypes.Identifiable{
				Id: "POKT-UnitTestNet",
			},
			GeoZone: &coreTypes.Identifiable{
				Id: "geozone",
			},
		},
		RelayPayload: &coreTypes.Relay_RestPayload{
			RestPayload: &coreTypes.RESTPayload{
				HttpPath:    "/v1/height",
				RequestType: coreTypes.RESTRequestType_RESTRequestTypeGET,
			},
		},
	}

	for _, editor := range editors {
		editor(relay)
	}

	return relay
}

type configModifier func(*configs.ServicerConfig)

func withPrivateKey(key string) func(*configs.ServicerConfig) {
	return func(cfg *configs.ServicerConfig) {
		cfg.PrivateKey = key
	}
}

func testServicerConfig(editors ...configModifier) *configs.ServicerConfig {
	config := configs.ServicerConfig{
		PrivateKey: testServicer1PrivateKey.String(),
		Services: map[string]*configs.ServiceConfig{
			"POKT-UnitTestNet": testServiceConfig1,
			"ETH-Goerli":       testServiceConfig1,
			"RESTful-service":  testServiceConfig1,
		},
	}

	for _, editor := range editors {
		editor(&config)
	}

	return &config
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
	runtimeMgrMock.EXPECT().GetConfig().Return(&configs.Config{Servicer: cfg}).AnyTimes()

	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(height).AnyTimes()

	persistenceReadContextMock := mockModules.NewMockPersistenceReadContext(ctrl)
	persistenceReadContextMock.EXPECT().Release().AnyTimes()
	persistenceReadContextMock.EXPECT().GetServicer(gomock.Any(), gomock.Any()).Return(testServicer1, nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetIntParam(typesUtil.AppSessionTokensMultiplierParamName, session.SessionHeight).
		Return(testAppsTokensMultiplier, nil).AnyTimes()

	persistenceLocalContextMock := mockModules.NewMockPersistenceLocalContext(ctrl)
	persistenceLocalContextMock.EXPECT().StoreServicedRelay(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	persistenceLocalContextMock.EXPECT().GetSessionTokensUsed(gomock.Any()).Return(big.NewInt(usedSessionTokens), nil).AnyTimes()

	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()
	persistenceMock.EXPECT().GetLocalContext().Return(persistenceLocalContextMock, nil).AnyTimes()

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
