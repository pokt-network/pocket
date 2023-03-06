package raintree

import (
	"github.com/golang/mock/gomock"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime/configs"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
)

func mockBus(ctrl *gomock.Controller) *mockModules.MockBus {
	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(nil).AnyTimes()
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(0)).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	runtimeMgrMock.EXPECT().GetConfig().Return(configs.NewDefaultConfig()).AnyTimes()
	return busMock
}

func mockPeerstoreProvider(ctrl *gomock.Controller, pstore sharedP2P.Peerstore) *mocksP2P.MockPeerstoreProvider {
	peerstoreProviderMock := mocksP2P.NewMockPeerstoreProvider(ctrl)
	peerstoreProviderMock.EXPECT().GetStakedPeerstoreAtHeight(gomock.Any()).Return(pstore, nil).AnyTimes()
	return peerstoreProviderMock
}

func mockCurrentHeightProvider(ctrl *gomock.Controller, height uint64) *mocksP2P.MockCurrentHeightProvider {
	currentHeightProviderMock := mocksP2P.NewMockCurrentHeightProvider(ctrl)
	currentHeightProviderMock.EXPECT().CurrentHeight().Return(height).AnyTimes()
	return currentHeightProviderMock
}
