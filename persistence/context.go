package persistence

import (
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"

	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
)

// type _ modules.PersistenceContext = &PersistenceContext{}

type PersistenceContext struct {
	kvstore kvstore.KVStore
}

func (p *PersistenceContext) Commit() error {
	return nil
}

func CreatePersistenceContext() (modules.PersistenceContext, error) {
	memKVStore := kvstore.NewMemKVStore()

	context := &PersistenceContext{
		kvstore: memKVStore,
	}

	ctrl := gomock.NewController(nil)
	persistenceContextMock := modulesMock.NewMockPersistenceContext(ctrl)

	persistenceContextMock.EXPECT().Commit().DoAndReturn(context.Commit).AnyTimes()
	// utilityContextMock := modulesMock.NewMockUtilityContext(ctrl)
	// persistenceContextMock := modulesMock.NewMockPersistenceContext(ctrl)

	// utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	// utilityMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	// utilityMock.EXPECT().
	// 	NewContext(gomock.Any()).
	// 	Return(utilityContextMock, nil).
	// 	AnyTimes()

	// utilityContextMock.EXPECT().GetPersistenceContext().Return(persistenceContextMock).AnyTimes()
	// utilityContextMock.EXPECT().ReleaseContext().Return().AnyTimes()
	// utilityContextMock.EXPECT().
	// 	GetProposalTransactions(gomock.Any(), maxTxBytes, gomock.AssignableToTypeOf(emptyByzValidators)).
	// 	Return(make([][]byte, 0), nil).
	// 	AnyTimes()
	// utilityContextMock.EXPECT().
	// 	ApplyProposalTransactions(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
	// 	Return(appHash, nil).
	// 	AnyTimes()

	//

	return persistenceContextMock, nil
}
