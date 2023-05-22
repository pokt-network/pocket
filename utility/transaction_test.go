package utility

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	mockPersistence "github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGetIndexedTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test data
	tx := types.Transaction{}
	txProtoBytes, err := proto.Marshal(&tx)
	require.NoError(t, err)
	// txString := string(txProtoBytes)

	tests := []struct {
		name            string
		txExists        bool
		txExistsErr     error
		getByHashErr    error
		expectErr       error
		expectIndexedTx bool
	}{
		{"returns indexed transaction when it exists", true, nil, nil, nil, true},
		{"returns error when transaction doesn't exist", false, nil, nil, error(types.ErrTransactionNotCommitted()), false},
		{"handles error from TransactionExists", false, fmt.Errorf("some error"), nil, fmt.Errorf("some error"), false},
		{"handles error from GetByHash", true, nil, fmt.Errorf("some error"), fmt.Errorf("some error"), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// mockPersistenceMod := mockModules.NewMockPersistenceModule(ctrl)
			// mockPersistenceMod.EXPECT().SetBus(gomock.Any()).AnyTimes()
			// mockPersistenceMod.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()

			// Preparing the environment
			_, utilityMod := prepareEnvironmentWithPersistenceMock(t, 0, 0, 0, 0, mockPersistenceMod)

			// Mock TransactionExists method
			// mockPersistenceMod.EXPECT().TransactionExists(gomock.Eq(txString)).Return(test.txExists, test.txExistsErr)
			mockPersistenceMod.EXPECT().TransactionExists(gomock.Any()).Return(test.txExists, test.txExistsErr)

			// Mock GetTxIndexer method
			mockIndexer := mockPersistence.NewMockTxIndexer(ctrl)
			mockIndexer.EXPECT().GetByHash(gomock.Any()).Return(&types.IndexedTransaction{}, test.getByHashErr)
			mockPersistenceMod.EXPECT().GetTxIndexer().Return(mockIndexer)

			idTx, err := utilityMod.GetIndexedTransaction(txProtoBytes)
			if test.expectErr != nil {
				fmt.Println("OLSH", test.expectErr, "\n", err)
				assert.ErrorIs(t, err, test.expectErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectIndexedTx, idTx != nil)
		})
	}
}
