package persistence_testutil

import (
	"fmt"
	"github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

// Persistence mock - only needed for validatorMap access
func BasePersistenceMock(t *testing.T, busMock *mock_modules.MockBus, genesisState *genesis.GenesisState) *mock_modules.MockPersistenceModule {
	ctrl := gomock.NewController(t)

	persistenceModuleMock := mock_modules.NewMockPersistenceModule(ctrl)
	readCtxMock := mock_modules.NewMockPersistenceReadContext(ctrl)

	readCtxMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	persistenceModuleMock.EXPECT().NewReadContext(gomock.Any()).Return(readCtxMock, nil).AnyTimes()
	readCtxMock.EXPECT().Release().AnyTimes()

	persistenceModuleMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	persistenceModuleMock.EXPECT().SetBus(busMock).AnyTimes()
	persistenceModuleMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	busMock.RegisterModule(persistenceModuleMock)

	return persistenceModuleMock
}

// Creates a persistence module mock with mock implementations of some basic functionality
func PersistenceMockWithBlockStore(t *testing.T, _ modules.EventsChannel, bus modules.Bus) *mock_modules.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := mock_modules.NewMockPersistenceModule(ctrl)
	persistenceReadContextMock := mock_modules.NewMockPersistenceReadContext(ctrl)

	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()

	persistenceMock.EXPECT().ReleaseWriteContext().Return(nil).AnyTimes()

	blockStoreMock := mock_kvstore.NewMockKVStore(ctrl)

	blockStoreMock.EXPECT().Get(gomock.Any()).DoAndReturn(func(height []byte) ([]byte, error) {
		heightInt := utils.HeightFromBytes(height)
		if bus.GetConsensusModule().CurrentHeight() < heightInt {
			return nil, fmt.Errorf("requested height is higher than current height of the node's consensus module")
		}
		blockWithHeight := &types.Block{
			BlockHeader: &types.BlockHeader{
				Height: utils.HeightFromBytes(height),
			},
		}
		return codec.GetCodec().Marshal(blockWithHeight)
	}).AnyTimes()

	persistenceMock.EXPECT().GetBlockStore().Return(blockStoreMock).AnyTimes()

	persistenceReadContextMock.EXPECT().GetMaximumBlockHeight().DoAndReturn(func() (uint64, error) {
		height := bus.GetConsensusModule().CurrentHeight()
		return height, nil
	}).AnyTimes()

	persistenceReadContextMock.EXPECT().GetMinimumBlockHeight().DoAndReturn(func() (uint64, error) {
		// mock minimum block height in persistence module to 1 if current height is equal or more than 1, else return 0 as the minimum height
		if bus.GetConsensusModule().CurrentHeight() >= 1 {
			return 1, nil
		}
		return 0, nil
	}).AnyTimes()

	persistenceReadContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(bus.GetRuntimeMgr().GetGenesis().Validators, nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()
	persistenceReadContextMock.EXPECT().Release().AnyTimes()

	return persistenceMock
}
