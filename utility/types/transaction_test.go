package types

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var (
	testingSenderPrivateKey, _ = crypto.GeneratePrivateKey()
	testingSenderPublicKey     = testingSenderPrivateKey.PublicKey()
	testingSenderAddr          = testingSenderPublicKey.Address()
	testingToAddr, _           = crypto.GenerateAddress()
)

func NewTestingMsg(_ *testing.T) Message {
	return &MessageSend{
		FromAddress: testingSenderAddr,
		ToAddress:   testingToAddr,
		Amount:      defaultAmount,
	}
}

func NewUnsignedTestingTransaction(t *testing.T) Transaction {
	codec := types.GetCodec()
	msg := NewTestingMsg(t)

	anyMsg, err := codec.ToAny(msg)
	require.NoError(t, err)

	return Transaction{
		Msg:   anyMsg,
		Nonce: types.BigIntToString(types.RandBigInt()),
	}
}

func TestTransactionBytesAndFromBytes(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	bz, err := tx.Bytes()
	require.NoError(t, err)

	tx2, err := TransactionFromBytes(bz)
	require.NoError(t, err)

	hash1, err := tx.Hash()
	require.NoError(t, err)

	hash2, err := tx2.Hash()
	require.NoError(t, err)

	require.Equal(t, hash1, hash2, "transaction hash mismatch")
	require.Equal(t, proto.Clone(&tx), proto.Clone(tx2), "transaction mismatch")
}

func TestTransaction_Message(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	msg, err := tx.Message()
	require.NoError(t, err)

	expected := NewTestingMsg(t)
	require.NotEqual(t, expected, msg)

	require.Equal(t, msg.ProtoReflect().Type(), expected.ProtoReflect().Type())
	message := msg.(*MessageSend)
	expectedMessage := expected.(*MessageSend)
	if message.Amount != expectedMessage.Amount ||
		!bytes.Equal(message.ToAddress, expectedMessage.ToAddress) ||
		!bytes.Equal(message.FromAddress, expectedMessage.FromAddress) {
		t.Fatal("unequal messages")
	}
}

func TestTransaction_Sign(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	err := tx.Sign(testingSenderPrivateKey)
	require.NoError(t, err)

	msg, err := tx.SignBytes()
	require.NoError(t, err)
	verified := testingSenderPublicKey.Verify(msg, tx.Signature.Signature)
	require.True(t, verified, "signature should be verified")
}

func TestTransaction_ValidateBasic(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	err := tx.Sign(testingSenderPrivateKey)
	require.NoError(t, err)

	er := tx.ValidateBasic()
	require.NoError(t, er)

	txNoNonce := proto.Clone(&tx).(*Transaction)
	txNoNonce.Nonce = ""
	er = txNoNonce.ValidateBasic()
	require.Equal(t, types.ErrEmptyNonce().Code(), er.Code())

	txInvalidMessageAny := proto.Clone(&tx).(*Transaction)
	txInvalidMessageAny.Msg = nil
	if err := txInvalidMessageAny.ValidateBasic(); err.Code() != types.ErrProtoFromAny(err).Code() {
		require.NoError(t, err)
	}

	txEmptySig := proto.Clone(&tx).(*Transaction)
	txEmptySig.Signature = nil
	er = txEmptySig.ValidateBasic()
	require.Equal(t, types.ErrEmptySignature().Code(), er.Code())

	txEmptyPublicKey := proto.Clone(&tx).(*Transaction)
	txEmptyPublicKey.Signature.PublicKey = nil
	er = txEmptyPublicKey.ValidateBasic()
	require.Equal(t, types.ErrEmptyPublicKey().Code(), er.Code())

	txInvalidPublicKey := proto.Clone(&tx).(*Transaction)
	txInvalidPublicKey.Signature.PublicKey = []byte("publickey")
	if err := txInvalidPublicKey.ValidateBasic(); err.Code() != types.ErrNewPublicKeyFromBytes(err).Code() {
		require.NoError(t, err)
	}

	txInvalidSignature := proto.Clone(&tx).(*Transaction)
	tx.Signature.PublicKey = testingSenderPublicKey.Bytes()
	txInvalidSignature.Signature.Signature = []byte("signature")
	er = txInvalidSignature.ValidateBasic()
	require.Equal(t, types.ErrSignatureVerificationFailed().Code(), er.Code())

}
