package network

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

var (
	testServiceUrl = "10.0.0.%d:8080"
)

func MockBus(ctrl *gomock.Controller) *mock_modules.MockBus {
	consensusMock := mock_modules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(0)).AnyTimes()

	runtimeMgrMock := mock_modules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(configs.NewDefaultConfig()).AnyTimes()

	busMock := mock_modules.NewMockBus(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(nil).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()

	return busMock
}

func MockAddrBookProvider(ctrl *gomock.Controller, addrBook types.AddrBook) *mock_types.MockAddrBookProvider {
	addrBookProviderMock := mock_types.NewMockAddrBookProvider(ctrl)
	addrBookProviderMock.EXPECT().GetStakedAddrBookAtHeight(gomock.Any()).Return(addrBook, nil).AnyTimes()
	return addrBookProviderMock
}

func MockCurrentHeightProvider(ctrl *gomock.Controller, height uint64) *mock_types.MockCurrentHeightProvider {
	currentHeightProviderMock := mock_types.NewMockCurrentHeightProvider(ctrl)
	currentHeightProviderMock.EXPECT().CurrentHeight().Return(height).AnyTimes()
	return currentHeightProviderMock
}

// Generates an address book with a random set of `n` addresses
func GetAddrBook(t *testing.T, n int) (addrBook types.AddrBook) {
	if n > 254 {
		panic("requires refactor to produce valid IPv4 addresses for n > 254")
	}

	addrBook = make([]*types.NetworkPeer, 0)
	for i := 0; i < n; i++ {
		pubKey, err := crypto.GeneratePublicKey()
		if t != nil {
			require.NoError(t, err)
		}
		addrBook = append(addrBook, &types.NetworkPeer{
			PublicKey:  pubKey,
			Address:    pubKey.Address(),
			ServiceUrl: fmt.Sprintf(testServiceUrl, i),
		})
	}
	return
}
