package servicer

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/runtime/configs"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

var testServicer1 = &coreTypes.Actor{
	ActorType: coreTypes.ActorType_ACTOR_TYPE_SERVICER,
	Address:   "a3d9ea9d9ad9c58bb96ec41340f83cb2cabb6496",
	PublicKey: "a6cd0a304c38d76271f74dd3c90325144425d904ef1b9a6fbab9b201d75a998b",
	Chains:    []string{"0021"},
}

func TestAdmitRelay(t *testing.T) {
	currentHeight := uint64(9)
	testCases := []struct {
		name     string
		relay    *coreTypes.Relay
		expected error
	}{
		{
			name:  "valid relay is admitted",
			relay: testRelay("0021", int64(currentHeight)),
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
			relay:    testRelay("foo", 8),
			expected: errValidateRelayMeta,
		},
		{
			name:     "Relay with height set in a past session is rejected",
			relay:    testRelay("0021", 5),
			expected: errValidateBlockHeight,
		},
		{
			name:     "Relay with height set in a future session is rejected",
			relay:    testRelay("0021", 9999),
			expected: errValidateBlockHeight,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			config := testServicerConfig()
			session := testSession(
				sessionNumber(2),
				sessionBlocks(4),
				sessionHeight(2),
				sessionServicers(testServicer1),
			)
			mockBus := mockBus(t, &config, currentHeight, session)

			servicerMod, err := CreateServicer(mockBus)
			require.NoError(t, err)

			err = servicerMod.(*servicer).admitRelay(testCase.relay)
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("Expected error %v got: %v", testCase.expected, err)
			}
		})
	}
}

func testRelay(chain string, height int64) *coreTypes.Relay {
	return &coreTypes.Relay{
		Meta: &coreTypes.RelayMeta{
			ServicerPublicKey: testServicer1.PublicKey,
			BlockHeight:       height,
			RelayChain: &coreTypes.Identifiable{
				Id: chain,
			},
			GeoZone: &coreTypes.Identifiable{
				Id: "geozone",
			},
		},
	}
}

func testServicerConfig() configs.ServicerConfig {
	return configs.ServicerConfig{
		PublicKey: testServicer1.PublicKey,
		Address:   testServicer1.Address,
		Chains:    testServicer1.Chains,
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
		Id: "session-1",
	}
	for _, editor := range editors {
		editor(&session)
	}
	return &session
}

// Create a mockBus with mock implementations of consensus and utility modules
func mockBus(t *testing.T, cfg *configs.ServicerConfig, height uint64, session *coreTypes.Session) *mockModules.MockBus {
	ctrl := gomock.NewController(t)
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(&configs.Config{Servicer: cfg}).AnyTimes()

	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(height).AnyTimes()

	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	persistenceReadContextMock := mockModules.NewMockPersistenceReadContext(ctrl)
	persistenceReadContextMock.EXPECT().GetServicer(gomock.Any(), gomock.Any()).Return(testServicer1, nil).AnyTimes()
	persistenceReadContextMock.EXPECT().Release().AnyTimes()

	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()

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
