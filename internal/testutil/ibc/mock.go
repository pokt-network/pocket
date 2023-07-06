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
	t.Helper()
	ctrl := gomock.NewController(t)
	ibcMock := mockModules.NewMockIBCModule(ctrl)

	ibcMock.EXPECT().Start().Return(nil).AnyTimes()
	ibcMock.EXPECT().SetBus(bus).Return().AnyTimes()
	ibcMock.EXPECT().GetBus().Return(bus).AnyTimes()
	ibcMock.EXPECT().GetModuleName().Return(modules.IBCModuleName).AnyTimes()

	return ibcMock
}

func IBCMockWithHost(t gocuke.TestingT, bus modules.Bus) (
	*mockModules.MockIBCModule,
	*mockModules.MockIBCHostSubmodule,
) {
	t.Helper()

	ibcMock := BaseIBCMock(t, bus)
	hostMock := BaseIBCHostMock(t, bus)

	return ibcMock, hostMock
}

// BaseIBCHostMock returns a mock IBC Host submodule
func BaseIBCHostMock(t gocuke.TestingT, bus modules.Bus) *mockModules.MockIBCHostSubmodule {
	t.Helper()
	ctrl := gomock.NewController(t)
	hostMock := mockModules.NewMockIBCHostSubmodule(ctrl)

	hostMock.EXPECT().SetBus(bus).Return().AnyTimes()
	hostMock.EXPECT().GetBus().Return(bus).AnyTimes()
	hostMock.EXPECT().GetModuleName().Return(modules.IBCHostSubmoduleName).AnyTimes()
	hostMock.EXPECT().GetTimestamp().DoAndReturn(func() uint64 {
		unix := time.Now().Unix()
		return uint64(unix)
	}).AnyTimes()

	prov := mockModules.NewMockProvableStore(ctrl)
	hostMock.EXPECT().GetProvableStore(prov).AnyTimes()

	bscMock := BaseBulkStoreCacherMock(t, bus)
	bus.RegisterModule(hostMock)
	bus.RegisterModule(bscMock)

	return hostMock
}

// BaseBulkStoreCacherMock returns a mock BulkStoreCacher submodule mock
func BaseBulkStoreCacherMock(t gocuke.TestingT, bus modules.Bus) *mockModules.MockBulkStoreCacher {
	t.Helper()
	ctrl := gomock.NewController(t)
	storeMock := mockModules.NewMockBulkStoreCacher(ctrl)
	provableStoreMock := mockModules.NewMockProvableStore(ctrl)

	storeMock.EXPECT().SetBus(bus).Return().AnyTimes()
	storeMock.EXPECT().GetBus().Return(bus).AnyTimes()
	storeMock.EXPECT().GetModuleName().Return(modules.BulkStoreCacherModuleName).AnyTimes()
	storeMock.EXPECT().AddStore(gomock.Any()).Return(nil).AnyTimes()
	storeMock.EXPECT().GetStore(gomock.Any()).Return(provableStoreMock, nil).AnyTimes()
	storeMock.EXPECT().RemoveStore(gomock.Any()).Return(nil).AnyTimes()
	storeMock.EXPECT().FlushAllEntries().Return(nil).AnyTimes()
	storeMock.EXPECT().PruneCaches(gomock.Any()).Return(nil).AnyTimes()
	storeMock.EXPECT().RestoreCaches().Return(nil).AnyTimes()

	return storeMock
}
