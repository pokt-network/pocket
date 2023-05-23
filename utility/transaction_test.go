package utility

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/core/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGetIndexedTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test data - Prepare Transaction
	emptyTx := types.Transaction{}
	txProtoBytes, err := proto.Marshal(&emptyTx)
	require.NoError(t, err)

	// Test data - Prepare IndexedTransaction
	idxTx := &coreTypes.IndexedTransaction{
		Tx:            txProtoBytes,
		Height:        0,
		Index:         0,
		ResultCode:    0,
		Error:         "Olshansky",
		SignerAddr:    "Olshansky",
		RecipientAddr: "Olshansky",
		MessageType:   "Olshansky",
	}

	// Prepare the environment
	_, utilityMod, persistenceMod := prepareEnvironment(t, 0, 0, 0, 0)

	// Index a test transaction
	txIndexer := persistenceMod.GetTxIndexer()
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)

	tests := []struct {
		name         string
		txProtoBytes []byte
		txExists     bool
		expectErr    error
	}{
		{"returns indexed transaction when it exists", txProtoBytes, true, nil},
		{"returns error when transaction doesn't exist", []byte("Does not exist"), false, types.ErrTransactionNotCommitted()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			idTx, err := utilityMod.GetIndexedTransaction(test.txProtoBytes)
			if test.expectErr != nil {
				require.EqualError(t, err, test.expectErr.Error())
				require.Nil(t, idTx)
			} else {
				require.NoError(t, err)
				require.NotNil(t, idTx)
			}
		})
	}
}
