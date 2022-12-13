package raintree

import (
	"github.com/golang/mock/gomock"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

func mockDummyBus(ctrl *gomock.Controller) *mockModules.MockBus {
	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(nil).AnyTimes()
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(0)).AnyTimes()
	consensusMock.EXPECT().ValidatorMap().Return(nil).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	return busMock
}

func mockAddrBookProvider(ctrl *gomock.Controller, addrBook typesP2P.AddrBook) *mocksP2P.MockAddrBookProvider {
	addrBookProviderMock := mocksP2P.NewMockAddrBookProvider(ctrl)
	addrBookProviderMock.EXPECT().ActorToAddrBook(gomock.Any()).Return(addrBook, nil).AnyTimes()
	addrBookProviderMock.EXPECT().GetStakedAddrBookAtHeight(gomock.Any()).Return(addrBook, nil).AnyTimes()
	return addrBookProviderMock
}
