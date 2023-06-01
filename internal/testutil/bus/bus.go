package bus_testutil

import (
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/runtime"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

func NewBus(
	t gocuke.TestingT,
	privKey crypto.PrivateKey,
	serviceURL string,
	genesisState *genesis.GenesisState,
	busEventHandlerFactory testutil.BusEventHandlerFactory,
) *mock_modules.MockBus {
	t.Helper()

	runtimeMgrMock := runtime_testutil.BaseRuntimeManagerMock(
		t, privKey,
		serviceURL,
		genesisState,
	)
	busMock := testutil.BusMockWithEventHandler(t, runtimeMgrMock, busEventHandlerFactory)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	return busMock
}
