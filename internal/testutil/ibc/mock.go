package ibc

import (
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/regen-network/gocuke"
)

// BaseIbcMock returns a mock IBC module without a Host
func BaseIbcMock(t gocuke.TestingT, busMock *mockModules.MockBus) *mockModules.MockIBCModule {
	ctrl := gomock.NewController(t)
	ibcMock := mockModules.NewMockIBCModule(ctrl)

	ibcMock.EXPECT().Start().Return(nil).AnyTimes()
	ibcMock.EXPECT().SetBus(busMock).Return().AnyTimes()
	ibcMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	ibcMock.EXPECT().GetModuleName().Return(modules.IBCModuleName).AnyTimes()

	return ibcMock
}

// IbcMockWithHost returns a mock IBC module with a Host
func IbcMockWithHost(t gocuke.TestingT, _ modules.EventsChannel) *mockModules.MockIBCModule {
	ctrl := gomock.NewController(t)
	ibcMock := mockModules.NewMockIBCModule(ctrl)

	ibcMock.EXPECT().Start().Return(nil).AnyTimes()
	ibcMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	ibcMock.EXPECT().GetModuleName().Return(modules.IBCModuleName).AnyTimes()

	hostMock := mockModules.NewMockIBCHost(ctrl)
	hostMock.EXPECT().GetTimestamp().DoAndReturn(func() uint64 {
		timestamp := time.Now().Unix()
		return uint64(timestamp)
	}).AnyTimes()

	ibcMock.EXPECT().GetHost().Return(hostMock).AnyTimes()

	return ibcMock
}
