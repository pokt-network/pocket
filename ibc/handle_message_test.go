package ibc

import (
	"testing"

	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/core/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestHandleMessage_ErrorAlreadyInMempool(t *testing.T) {
	// Prepare test data
	_, tx := prepareUpdateMessage(t, []byte("key"), []byte("value"))
	txProtoBytes, err := codec.GetCodec().Marshal(tx)
	require.NoError(t, err)

	// Prepare the environment
	_, utilityMod, _, _ := prepareEnvironment(t, 1, 0, 0, 0)

	// Manually add the tx to the mempool
	err = utilityMod.GetMempool().AddTx(txProtoBytes)
	require.NoError(t, err)

	// Error on having a duplciate transaction
	err = utilityMod.HandleTransaction(txProtoBytes)
	require.Error(t, err)
	require.EqualError(t, err, coreTypes.ErrDuplicateTransaction().Error())
}

func TestHandleMessage_ErrorAlreadyCommitted(t *testing.T) {
	// Prepare the environment
	_, utilityMod, persistenceMod, _ := prepareEnvironment(t, 0, 0, 0, 0)
	idxTx := prepareIndexedMessage(t, persistenceMod.GetTxIndexer())

	// Error on having an indexed transaction
	err := utilityMod.HandleTransaction(idxTx.Tx)
	require.Error(t, err)
	require.EqualError(t, err, coreTypes.ErrTransactionAlreadyCommitted().Error())
}

func TestHandleMessage_BasicValidation_Message(t *testing.T) {
	updateMsg, _ := prepareUpdateMessage(t, []byte("key"), []byte("value"))
	require.NoError(t, updateMsg.ValidateBasic())
	pruneMsg, _ := preparePruneMessage(t, []byte("key"))
	require.NoError(t, pruneMsg.ValidateBasic())

	testCases := []struct {
		name     string
		msg      *ibcTypes.IbcMessage
		expected error
	}{
		{
			name:     "Invalid Update Message: Empty Key",
			msg:      CreateUpdateStoreMessage(nil, []byte("value")),
			expected: coreTypes.ErrNilField("key"),
		},
		{
			name:     "Invalid Update Message: Empty Value",
			msg:      CreateUpdateStoreMessage([]byte("key"), nil),
			expected: coreTypes.ErrNilField("value"),
		},
		{
			name:     "Invalid Prune Message: Empty Key",
			msg:      CreatePruneStoreMessage(nil),
			expected: coreTypes.ErrNilField("key"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			require.EqualError(t, err, tc.expected.Error())
		})
	}
}

func TestHandleMessage_BasicValidation_Transaction(t *testing.T) {
	// Prepare the environment
	_, utilityMod, _, _ := prepareEnvironment(t, 1, 0, 0, 0)

	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	_, validTx := prepareUpdateMessage(t, []byte("key"), []byte("value"))
	require.NoError(t, err)
	err = validTx.Sign(privKey)
	require.NoError(t, err)

	txProtoBytes, err := codec.GetCodec().Marshal(validTx)
	require.NoError(t, err)

	err = utilityMod.HandleTransaction(txProtoBytes)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		msg      *ibcTypes.IbcMessage
		expected error
	}{
		{
			name:     "Invalid Update Message: Empty Key",
			msg:      CreateUpdateStoreMessage(nil, []byte("value")),
			expected: coreTypes.ErrEmptySignatureStructure(),
		},
		{
			name:     "Invalid Update Message: Empty Value",
			msg:      CreateUpdateStoreMessage([]byte("key"), nil),
			expected: coreTypes.ErrEmptySignatureStructure(),
		},
		{
			name:     "Invalid Prune Message: Empty Key",
			msg:      CreatePruneStoreMessage(nil),
			expected: coreTypes.ErrEmptySignatureStructure(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := ConvertIBCMessageToTx(tc.msg)
			require.NoError(t, err)
			txProtoBytes, err := codec.GetCodec().Marshal(tx)
			require.NoError(t, err)
			err = utilityMod.HandleTransaction(txProtoBytes)
			require.EqualError(t, err, tc.expected.Error())
		})
	}
}

func TestHandleMessage_GetIndexedMessage(t *testing.T) {
	// Prepare the environment
	_, utilityMod, persistenceMod, _ := prepareEnvironment(t, 1, 0, 0, 0)
	idxTx := prepareIndexedMessage(t, persistenceMod.GetTxIndexer())

	tests := []struct {
		name         string
		txProtoBytes []byte
		txExists     bool
		expectErr    error
	}{
		{"returns indexed transaction when it exists", idxTx.Tx, true, nil},
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

func TestHandleMessage_AddToMempool(t *testing.T) {
	// prepare the environment
	_, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	require.Len(t, ibcMod.GetBus().GetUtilityModule().GetMempool().GetAll(), 0)
	msg, _ := prepareUpdateMessage(t, []byte("key"), []byte("value"))
	anyMsg, err := codec.GetCodec().ToAny(msg)
	require.NoError(t, err)
	require.NoError(t, ibcMod.HandleMessage(anyMsg))
	require.Len(t, ibcMod.GetBus().GetUtilityModule().GetMempool().GetAll(), 1)
}

func prepareUpdateMessage(t *testing.T, key, value []byte) (*ibcTypes.IbcMessage, *coreTypes.Transaction) {
	t.Helper()
	msg := CreateUpdateStoreMessage(key, value)
	tx, err := ConvertIBCMessageToTx(msg)
	require.NoError(t, err)
	return msg, tx
}

func preparePruneMessage(t *testing.T, key []byte) (*ibcTypes.IbcMessage, *coreTypes.Transaction) {
	t.Helper()
	msg := CreatePruneStoreMessage(key)
	tx, err := ConvertIBCMessageToTx(msg)
	require.NoError(t, err)
	return msg, tx
}

func prepareIndexedMessage(t *testing.T, txIndexer indexer.TxIndexer) *coreTypes.IndexedTransaction {
	t.Helper()
	_, tx := preparePruneMessage(t, []byte{})
	txProtoBytes, err := codec.GetCodec().Marshal(tx)
	require.NoError(t, err)

	// Test data - Prepare IndexedTransaction
	idxTx := &coreTypes.IndexedTransaction{
		Tx:            txProtoBytes,
		Height:        0,
		Index:         0,
		ResultCode:    0,
		Error:         "h5law",
		SignerAddr:    "h5law",
		RecipientAddr: "h5law",
		MessageType:   "h5law",
	}

	// Index a test transaction
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)

	return idxTx
}
