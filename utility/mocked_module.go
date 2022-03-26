package utility

import (
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
)

var maxTxBytes = 90000
var emptyByzValidators = make([][]byte, 0)
var appHash []byte

func CreateMockedModule(_ *config.Config) (modules.UtilityModule, error) {
	ctrl := gomock.NewController(nil)
	utilityMock := modulesMock.NewMockUtilityModule(ctrl)
	utilityContextMock := modulesMock.NewMockUtilityContext(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceContext(ctrl)

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
