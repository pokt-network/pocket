package ibc

import (
	"errors"
	"fmt"
	"testing"

	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/stretchr/testify/require"
)

func TestHandleMessage_ErrorAlreadyInMempool(t *testing.T) {
	// Prepare test data
	_, tx := prepareUpdateMessage(t, []byte("key"), []byte("value"))
	txProtoBytes, err := codec.GetCodec().Marshal(tx)
	require.NoError(t, err)

	// Prepare the environment
	_, _, utilityMod, _, _ := prepareEnvironment(t, 1, 0, 0, 0)

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
	_, _, utilityMod, persistenceMod, _ := prepareEnvironment(t, 0, 0, 0, 0)
	idxTx := prepareIndexedMessage(t, persistenceMod.GetTxIndexer())

	// Error on having an indexed transaction
	err := utilityMod.HandleTransaction(idxTx.Tx)
	require.Error(t, err)
	require.EqualError(t, err, coreTypes.ErrTransactionAlreadyCommitted().Error())
}

func TestHandleMessage_BasicValidation_Message(t *testing.T) {
	testCases := []struct {
		name     string
		msg      *ibcTypes.IBCMessage
		expected error
	}{
		{
			name:     "Valid Update Message",
			msg:      ibcTypes.CreateUpdateStoreMessage([]byte("key"), []byte("value")),
			expected: nil,
		},
		{
			name:     "Valid Prune Message",
			msg:      ibcTypes.CreatePruneStoreMessage([]byte("key")),
			expected: nil,
		},
		{
			name:     "Invalid Update Message: Empty Key",
			msg:      ibcTypes.CreateUpdateStoreMessage(nil, []byte("value")),
			expected: coreTypes.ErrNilField("key"),
		},
		{
			name:     "Invalid Update Message: Empty Value",
			msg:      ibcTypes.CreateUpdateStoreMessage([]byte("key"), nil),
			expected: coreTypes.ErrNilField("value"),
		},
		{
			name:     "Invalid Prune Message: Empty Key",
			msg:      ibcTypes.CreatePruneStoreMessage(nil),
			expected: coreTypes.ErrNilField("key"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expected != nil {
				require.EqualError(t, err, tc.expected.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleMessage_BasicValidation_Transaction(t *testing.T) {
	// Prepare the environment
	_, _, utilityMod, _, _ := prepareEnvironment(t, 1, 0, 0, 0)

	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	falseKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	validUpdateMsg, validUpdateTx := prepareUpdateMessage(t, []byte("key"), []byte("value"))
	require.NoError(t, err)
	err = validUpdateTx.Sign(privKey)
	require.NoError(t, err)
	updateAny, err := codec.GetCodec().ToAny(validUpdateMsg.GetUpdate())
	require.NoError(t, err)
	bz, err := validUpdateTx.SignableBytes()
	require.NoError(t, err)
	falseUpdateSig, err := falseKey.Sign(bz)
	require.NoError(t, err)

	validPruneMsg, validPruneTx := preparePruneMessage(t, []byte("key"))
	require.NoError(t, err)
	err = validPruneTx.Sign(privKey)
	require.NoError(t, err)
	pruneAny, err := codec.GetCodec().ToAny(validPruneMsg.GetPrune())
	require.NoError(t, err)
	bz, err = validPruneTx.SignableBytes()
	require.NoError(t, err)
	falsePruneSig, err := falseKey.Sign(bz)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		tx       *coreTypes.Transaction
		expected error
	}{
		{
			name:     "Valid Update Transaction",
			tx:       validUpdateTx,
			expected: nil,
		},
		{
			name:     "Valid Prune Transaction",
			tx:       validPruneTx,
			expected: nil,
		},
		{
			name: "Invalid Update Transaction: Empty Nonce",
			tx: &coreTypes.Transaction{
				Msg: updateAny,
			},
			expected: coreTypes.ErrEmptyNonce(),
		},
		{
			name: "Invalid Prune Transaction: Empty Nonce",
			tx: &coreTypes.Transaction{
				Msg: pruneAny,
			},
			expected: coreTypes.ErrEmptyNonce(),
		},
		{
			name: "Invalid Update Transaction: Empty Signature",
			tx: &coreTypes.Transaction{
				Msg:   updateAny,
				Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
			},
			expected: coreTypes.ErrEmptySignatureStructure(),
		},
		{
			name: "Invalid Prune Transaction: Empty Signature",
			tx: &coreTypes.Transaction{
				Msg:   pruneAny,
				Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
			},
			expected: coreTypes.ErrEmptySignatureStructure(),
		},
		{
			name: "Invalid Update Transaction: Bad Key",
			tx: &coreTypes.Transaction{
				Msg:   updateAny,
				Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
				Signature: &coreTypes.Signature{
					PublicKey: []byte("bad key"),
					Signature: falsePruneSig,
				},
			},
			expected: coreTypes.ErrNewPublicKeyFromBytes(errors.New("the public key length is not valid, expected length 32, actual length: 7")),
		},
		{
			name: "Invalid Prune Transaction: Bad Signature",
			tx: &coreTypes.Transaction{
				Msg:   pruneAny,
				Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
				Signature: &coreTypes.Signature{
					PublicKey: []byte("bad key"),
					Signature: falsePruneSig,
				},
			},
			expected: coreTypes.ErrNewPublicKeyFromBytes(errors.New("the public key length is not valid, expected length 32, actual length: 7")),
		},
		{
			name: "Invalid Update Transaction: Bad Signature",
			tx: &coreTypes.Transaction{
				Msg:   updateAny,
				Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
				Signature: &coreTypes.Signature{
					PublicKey: privKey.PublicKey().Bytes(),
					Signature: falseUpdateSig,
				},
			},
			expected: coreTypes.ErrSignatureVerificationFailed(),
		},
		{
			name: "Invalid Prune Transaction: Bad Key",
			tx: &coreTypes.Transaction{
				Msg:   pruneAny,
				Nonce: fmt.Sprintf("%d", crypto.GetNonce()),
				Signature: &coreTypes.Signature{
					PublicKey: privKey.PublicKey().Bytes(),
					Signature: falsePruneSig,
				},
			},
			expected: coreTypes.ErrSignatureVerificationFailed(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txProtoBytes, err := codec.GetCodec().Marshal(tc.tx)
			require.NoError(t, err)
			err = utilityMod.HandleTransaction(txProtoBytes)
			if tc.expected != nil {
				require.EqualError(t, err, tc.expected.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleMessage_GetIndexedMessage(t *testing.T) {
	// Prepare the environment
	_, _, utilityMod, persistenceMod, _ := prepareEnvironment(t, 1, 0, 0, 0)
	idxTx := prepareIndexedMessage(t, persistenceMod.GetTxIndexer())

	tests := []struct {
		name         string
		txProtoBytes []byte
		txExists     bool
		expectErr    error
	}{
		{"returns indexed transaction when it exists", idxTx.Tx, true, nil},
		{"returns error when transaction doesn't exist", []byte("Does not exist"), false, coreTypes.ErrTransactionNotCommitted()},
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

func prepareUpdateMessage(t *testing.T, key, value []byte) (*ibcTypes.IBCMessage, *coreTypes.Transaction) {
	t.Helper()
	msg := ibcTypes.CreateUpdateStoreMessage(key, value)
	letter, err := messaging.PackMessage(msg)
	require.NoError(t, err)
	t.Log(letter.GetContent().GetTypeUrl())
	tx, err := ibcTypes.ConvertIBCMessageToTx(msg)
	require.NoError(t, err)
	return msg, tx
}

func preparePruneMessage(t *testing.T, key []byte) (*ibcTypes.IBCMessage, *coreTypes.Transaction) {
	t.Helper()
	msg := ibcTypes.CreatePruneStoreMessage(key)
	tx, err := ibcTypes.ConvertIBCMessageToTx(msg)
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
