package utility

import (
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
)

var maxTxBytes = 90000
var emptyByzValidators = make([][]byte, 0)
var appHash []byte

// var emptyTxs = make([][]byte, 0)

func CreateMockedModule(_ *config.Config) (modules.UtilityModule, error) {
	ctrl := gomock.NewController(nil)
	utilityMock := mock_modules.NewMockUtilityModule(ctrl)
	utilityContextMock := mock_modules.NewMockUtilityContext(ctrl)
	persistenceContextMock := mock_modules.NewMockPersistenceContext(ctrl)

	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	utilityMock.EXPECT().
		NewContext(gomock.Any()).
		Return(utilityContextMock, nil).
		AnyTimes()

	utilityContextMock.EXPECT().GetPersistanceContext().Return(persistenceContextMock).AnyTimes()
	utilityContextMock.EXPECT().ReleaseContext().Return().AnyTimes()
	utilityContextMock.EXPECT().
		GetTransactionsForProposal(gomock.Any(), maxTxBytes, gomock.AssignableToTypeOf(emptyByzValidators)).
		Return(make([][]byte, 0), nil).
		AnyTimes()
	utilityContextMock.EXPECT().
		ApplyBlock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(appHash, nil).
		AnyTimes()

	persistenceContextMock.EXPECT().Commit().Return(nil).AnyTimes()

	return utilityMock, nil
}
