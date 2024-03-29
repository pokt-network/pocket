package utility

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/codec"
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	util_types "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleTransaction_ErrorAlreadyInMempool(t *testing.T) {
	// Prepare test data
	emptyTx := core_types.Transaction{}
	txProtoBytes, err := proto.Marshal(&emptyTx)
	require.NoError(t, err)

	// Prepare the environment
	_, utilityMod, _ := prepareEnvironment(t, 0, 0, 0, 0)

	// Manually add the tx to the mempool
	err = utilityMod.GetMempool().AddTx(txProtoBytes)
	require.NoError(t, err)

	// Error on having a duplciate transaction
	err = utilityMod.HandleTransaction(txProtoBytes)
	require.Error(t, err)
	require.EqualError(t, err, core_types.ErrDuplicateTransaction().Error())
}

func TestHandleTransaction_ErrorAlreadyCommitted(t *testing.T) {
	// Prepare the environment
	_, utilityMod, persistenceMod := prepareEnvironment(t, 0, 0, 0, 0)

	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	emptyTx := &core_types.Transaction{}
	err = emptyTx.Sign(privKey)
	require.NoError(t, err)
	txProtoBytes, err := codec.GetCodec().Marshal(emptyTx)
	require.NoError(t, err)

	// Test data - Prepare IndexedTransaction
	idxTx := &core_types.IndexedTransaction{
		Tx:            txProtoBytes,
		Height:        0,
		Index:         0,
		ResultCode:    0,
		Error:         "Olshansky",
		SignerAddr:    "Olshansky",
		RecipientAddr: "Olshansky",
		MessageType:   "Olshansky",
	}

	// Index a test transaction
	err = persistenceMod.GetTxIndexer().Index(idxTx)
	require.NoError(t, err)

	rwCtx, err := persistenceMod.NewRWContext(0)
	require.NoError(t, err)
	_, err = rwCtx.ComputeStateHash()
	require.NoError(t, err)
	rwCtx.Release()

	// Error on having an indexed transaction
	err = utilityMod.HandleTransaction(idxTx.Tx)
	require.Error(t, err)
	require.EqualError(t, err, core_types.ErrTransactionAlreadyCommitted().Error())
}

func TestHandleTransaction_BasicValidation(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	pubKey := privKey.PublicKey()

	message := &util_types.MessageSend{
		FromAddress: []byte("from"),
		ToAddress:   []byte("to"),
		Amount:      "10",
	}
	anyMessage, err := codec.GetCodec().ToAny(message)
	require.NoError(t, err)

	validTx := &core_types.Transaction{
		Nonce: strconv.Itoa(int(crypto.GetNonce())),
		Signature: &core_types.Signature{
			PublicKey: []byte("public key"),
			Signature: []byte("signature"),
		},
		Msg: anyMessage,
	}
	err = validTx.Sign(privKey)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		txProto     *core_types.Transaction
		expectedErr error
	}{
		{
			name:        "Invalid transaction: Missing Nonce",
			txProto:     &core_types.Transaction{},
			expectedErr: core_types.ErrEmptyNonce(),
		},
		{
			name: "Invalid transaction: Missing Signature Structure",
			txProto: &core_types.Transaction{
				Nonce: strconv.Itoa(int(crypto.GetNonce())),
			},
			expectedErr: core_types.ErrEmptySignatureStructure(),
		},
		{
			name: "Invalid transaction: Missing Signature",
			txProto: &core_types.Transaction{
				Nonce: strconv.Itoa(int(crypto.GetNonce())),
				Signature: &core_types.Signature{
					PublicKey: nil,
					Signature: nil,
				},
			},
			expectedErr: core_types.ErrEmptySignature(),
		},
		{
			name: "Invalid transaction: Missing Public Key",
			txProto: &core_types.Transaction{
				Nonce: strconv.Itoa(int(crypto.GetNonce())),
				Signature: &core_types.Signature{
					PublicKey: nil,
					Signature: []byte("bytes in place for signature but not actually valid"),
				},
			},
			expectedErr: core_types.ErrEmptyPublicKey(),
		},
		{
			name: "Invalid transaction: Invalid Public Key",
			txProto: &core_types.Transaction{
				Nonce: strconv.Itoa(int(crypto.GetNonce())),
				Signature: &core_types.Signature{
					PublicKey: []byte("invalid pub key"),
					Signature: []byte("bytes in place for signature but not actually valid"),
				},
			},
			expectedErr: core_types.ErrNewPublicKeyFromBytes(fmt.Errorf("the public key length is not valid, expected length 32, actual length: 15")),
		},
		// TODO(olshansky): Figure out why sometimes we do and don't need `\u00a0` in the error
		{
			name: "Invalid transaction: Invalid Message",
			txProto: &core_types.Transaction{
				Nonce: strconv.Itoa(int(crypto.GetNonce())),
				Signature: &core_types.Signature{
					PublicKey: pubKey.Bytes(),
					Signature: []byte("bytes in place for signature but not actually valid"),
				},
				Msg: nil,
			},
			expectedErr: core_types.ErrDecodeMessage(fmt.Errorf("proto: invalid empty type URL")),
		},
		{
			name: "Invalid transaction: Invalid Signature",
			txProto: &core_types.Transaction{
				Nonce: strconv.Itoa(int(crypto.GetNonce())),
				Signature: &core_types.Signature{
					PublicKey: pubKey.Bytes(),
					Signature: []byte("invalid signature"),
				},
				Msg: anyMessage,
			},
			expectedErr: core_types.ErrSignatureVerificationFailed(),
		},
		{
			name:        "Valid well-formatted transaction with valid signature",
			txProto:     validTx,
			expectedErr: nil,
		},
	}

	// Prepare the environment
	_, utilityMod, _ := prepareEnvironment(t, 0, 0, 0, 0)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txProtoBytes, err := codec.GetCodec().Marshal(tc.txProto)
			require.NoError(t, err)

			err = utilityMod.HandleTransaction(txProtoBytes)
			if tc.expectedErr != nil {
				errMsg := err.Error()
				errMsg = strings.Replace(errMsg, string('\u00a0'), " ", 1)
				require.EqualError(t, tc.expectedErr, errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetIndexedTransaction(t *testing.T) {
	// Prepare the environment
	_, utilityMod, persistenceMod := prepareEnvironment(t, 0, 0, 0, 0)
	idxTx := prepareEmptyIndexedTransaction(t, persistenceMod.GetTxIndexer())

	tests := []struct {
		name         string
		txProtoBytes []byte
		txExists     bool
		expectErr    error
	}{
		{"returns indexed transaction when it exists", idxTx.Tx, true, nil},
		{"returns error when transaction doesn't exist", []byte("Does not exist"), false, core_types.ErrTransactionNotCommitted()},
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

func prepareEmptyIndexedTransaction(t *testing.T, txIndexer indexer.TxIndexer) *core_types.IndexedTransaction {
	t.Helper()

	// Test data - Prepare Transaction
	emptyTx := core_types.Transaction{}
	txProtoBytes, err := proto.Marshal(&emptyTx)
	require.NoError(t, err)

	// Test data - Prepare IndexedTransaction
	idxTx := &core_types.IndexedTransaction{
		Tx:            txProtoBytes,
		Height:        0,
		Index:         0,
		ResultCode:    0,
		Error:         "Olshansky",
		SignerAddr:    "Olshansky",
		RecipientAddr: "Olshansky",
		MessageType:   "Olshansky",
	}

	// Index a test transaction
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)

	return idxTx
}
