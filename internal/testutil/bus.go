package testutil

import (
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/regen-network/gocuke"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

type BusEventHandler func(*messaging.PocketEnvelope)
type BusEventHandlerFactory func(t gocuke.TestingT, bus modules.Bus) BusEventHandler

// MinimalBusMock returns a bus mock with a module registry and minimal
// expectations registered to maximize re-usability.
func MinimalBusMock(
	t gocuke.TestingT,
	runtimeMgr modules.RuntimeMgr,
) *mock_modules.MockBus {
	t.Helper()

	ctrl := gomock.NewController(t)
	busMock := mock_modules.NewMockBus(ctrl)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgr).AnyTimes()
	busMock.EXPECT().RegisterModule(gomock.Any()).DoAndReturn(func(m modules.Module) {
		m.SetBus(busMock)
	}).AnyTimes()

	mockModulesRegistry := mock_modules.NewMockModulesRegistry(ctrl)

	// TODO_THIS_COMMIT: refactor - this doesn't belong here
	mockModulesRegistry.EXPECT().GetModule(peerstore_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(peerstore_provider.ModuleName)).AnyTimes()
	mockModulesRegistry.EXPECT().GetModule(current_height_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(current_height_provider.ModuleName)).AnyTimes()

	busMock.EXPECT().GetModulesRegistry().Return(mockModulesRegistry).AnyTimes()
	return busMock
}

// BaseBusMock returns a base bus mock which will accept any event,
// passing it to the provided handler function, any number of times.
func BaseBusMock(
	t gocuke.TestingT,
	runtimeMgr modules.RuntimeMgr,
) *mock_modules.MockBus {
	t.Helper()

	return WithoutBusEventHandler(t, MinimalBusMock(t, runtimeMgr))
}

// BusMockWithEventHandler returns a base bus mock which will accept any event,
// any number of times, calling the `handler` returned from `handlerFactory`
// with the event as an argument.
func BusMockWithEventHandler(
	t gocuke.TestingT,
	runtimeMgr modules.RuntimeMgr,
	handlerFactory BusEventHandlerFactory,
) *mock_modules.MockBus {
	t.Helper()

	busMock := MinimalBusMock(t, runtimeMgr)
	return WithBusEventHandler(t, busMock, handlerFactory)
}

// WithBusEventHandler adds an expectation to a bus mock such that it will accept
// any event, any number of times, calling the `handler` returned from `handlerFactory`
// with the event as an argument.
func WithBusEventHandler(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
	handlerFactory BusEventHandlerFactory,
) *mock_modules.MockBus {
	t.Helper()

	handler := handlerFactory(t, busMock)
	busMock.EXPECT().PublishEventToBus(gomock.Any()).Do(handler).AnyTimes()
	return busMock
}

// WithoutBusEventHandler adds an expectation to a bus mock such that it will accept
// any event, any number of times.
func WithoutBusEventHandler(
	t gocuke.TestingT,
	busMock *mock_modules.MockBus,
) *mock_modules.MockBus {
	t.Helper()

	busMock.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	return busMock
}
