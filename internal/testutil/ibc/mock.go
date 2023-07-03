package ibc

import (
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/regen-network/gocuke"
)

// BaseIBCMock returns a mock IBC module without a Host
func BaseIBCMock(t gocuke.TestingT, bus modules.Bus) *mockModules.MockIBCModule {
	ctrl := gomock.NewController(t)
	ibcMock := mockModules.NewMockIBCModule(ctrl)

	ibcMock.EXPECT().Start().Return(nil).AnyTimes()
	ibcMock.EXPECT().SetBus(bus).Return().AnyTimes()
	ibcMock.EXPECT().GetBus().Return(bus).AnyTimes()
	ibcMock.EXPECT().GetModuleName().Return(modules.IBCModuleName).AnyTimes()

	return ibcMock
}

// BaseIBCHostMock returns a mock IBC Host submodule
func BaseIBCHostMock(t gocuke.TestingT, busMock *mockModules.MockBus) *mockModules.MockIBCHostModule {
	ctrl := gomock.NewController(t)
	hostMock := mockModules.NewMockIBCHostModule(ctrl)

	hostMock.EXPECT().SetBus(busMock).Return().AnyTimes()
	hostMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	hostMock.EXPECT().GetModuleName().Return(modules.IBCHostModuleName).AnyTimes()
	hostMock.EXPECT().GetTimestamp().DoAndReturn(func() uint64 {
		unix := time.Now().Unix()
		return uint64(unix)
	})

	prov := mockModules.NewMockProvableStore(ctrl)
	hostMock.EXPECT().GetProvableStore(prov).AnyTimes()

	return hostMock
}
