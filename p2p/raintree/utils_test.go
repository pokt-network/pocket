package raintree

import (
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// TECHDEBT(#796): refactor/de-dup & separate definitions of mocks from one another.
func mockBus(ctrl *gomock.Controller, pstore typesP2P.Peerstore) *mockModules.MockBus {
	if pstore == nil {
		pstore = &typesP2P.PeerAddrMap{}
	}

	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().RegisterModule(gomock.Any()).DoAndReturn(func(m modules.Submodule) {
		m.SetBus(busMock)
	}).AnyTimes()
	busMock.EXPECT().GetPersistenceModule().Return(nil).AnyTimes()
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(0)).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()
	runtimeMgrMock.EXPECT().GetConfig().Return(configs.NewDefaultConfig()).AnyTimes()

	currentHeightProviderMock := mockModules.NewMockCurrentHeightProvider(ctrl)
	currentHeightProviderMock.EXPECT().CurrentHeight().Return(uint64(0)).AnyTimes()
	busMock.EXPECT().
		GetCurrentHeightProvider().
		Return(currentHeightProviderMock).
		AnyTimes()

	peerstoreProviderMock := mocksP2P.NewMockPeerstoreProvider(ctrl)
	peerstoreProviderMock.EXPECT().
		GetStakedPeerstoreAtHeight(gomock.Any()).
		Return(pstore, nil).
		AnyTimes()

	modulesRegistryMock := mockModules.NewMockModulesRegistry(ctrl)
	modulesRegistryMock.EXPECT().
		GetModule(gomock.Eq(peerstore_provider.PeerstoreProviderSubmoduleName)).
		Return(peerstoreProviderMock, nil).
		AnyTimes()
	modulesRegistryMock.EXPECT().
		GetModule(gomock.Eq(modules.CurrentHeightProviderSubmoduleName)).
		Return(currentHeightProviderMock, nil).
		AnyTimes()
	busMock.EXPECT().GetModulesRegistry().Return(modulesRegistryMock).AnyTimes()

	return busMock
}

func mockPeerstoreProvider(
	ctrl *gomock.Controller,
	pstore typesP2P.Peerstore,
) *mocksP2P.MockPeerstoreProvider {
	peerstoreProviderMock := mocksP2P.NewMockPeerstoreProvider(ctrl)
	peerstoreProviderMock.EXPECT().SetBus(gomock.Any()).AnyTimes()
	peerstoreProviderMock.EXPECT().GetBus().AnyTimes()
	peerstoreProviderMock.EXPECT().
		GetStakedPeerstoreAtHeight(gomock.Any()).
		Return(pstore, nil).
		AnyTimes()
	peerstoreProviderMock.EXPECT().
		GetModuleName().
		Return(peerstore_provider.PeerstoreProviderSubmoduleName).
		AnyTimes()
	return peerstoreProviderMock
}

func mockCurrentHeightProvider(ctrl *gomock.Controller, height uint64) *mockModules.MockCurrentHeightProvider {
	currentHeightProviderMock := mockModules.NewMockCurrentHeightProvider(ctrl)
	currentHeightProviderMock.EXPECT().CurrentHeight().Return(height).AnyTimes()
	return currentHeightProviderMock
}
