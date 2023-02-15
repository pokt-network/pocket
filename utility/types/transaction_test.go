package types

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

var (
	testingSenderPrivateKey, _ = crypto.GeneratePrivateKey()
	testingSenderPublicKey     = testingSenderPrivateKey.PublicKey()
	testingSenderAddr          = testingSenderPublicKey.Address()
	testingToAddr, _           = crypto.GenerateAddress()
)

func TestTransaction_BytesAndFromBytes(t *testing.T) {
	tx := newUnsignedTestingTransaction(t)
	bz, err := tx.Bytes()
	require.NoError(t, err)

	tx2, err := TxFromBytes(bz)
	require.NoError(t, err)

	hash1, err := tx.Hash()
	require.NoError(t, err)

	hash2, err := tx2.Hash()
	require.NoError(t, err)

	require.Equal(t, hash1, hash2, "transaction hash mismatch")
	require.Equal(t, proto.Clone(&tx), proto.Clone(tx2), "transaction mismatch")
}

func TestTransaction_GetMessage(t *testing.T) {
	tx := newUnsignedTestingTransaction(t)
	msg, err := tx.GetMessage()
	require.NoError(t, err)

	expected := newTestingMsgSend(t)
	require.NotEqual(t, expected, msg)
	require.Equal(t, msg.ProtoReflect().Type(), expected.ProtoReflect().Type())

	messageSend := msg.(*MessageSend)
	expectedMessageSend := expected.(*MessageSend)
	require.Equal(t, messageSend.Amount, expectedMessageSend.Amount, "unequal messages")
	require.Equal(t, messageSend.FromAddress, expectedMessageSend.FromAddress, "unequal messages")
	require.Equal(t, messageSend.ToAddress, expectedMessageSend.ToAddress, "unequal messages")
}

func TestTransaction_Sign(t *testing.T) {
	tx := newUnsignedTestingTransaction(t)

	err := tx.Sign(testingSenderPrivateKey)
	require.NoError(t, err)

	msg, er := tx.SignableBytes()
	require.NoError(t, er)

	verified := testingSenderPublicKey.Verify(msg, tx.Signature.Signature)
	require.True(t, verified, "signature should be verified")
}

func TestTransaction_ValidateBasic(t *testing.T) {
	tx := newUnsignedTestingTransaction(t)
	err := tx.Sign(testingSenderPrivateKey)
	require.NoError(t, err)

	er := tx.ValidateBasic()
	require.NoError(t, er)

	txNoNonce := proto.Clone(&tx).(*Transaction)
	txNoNonce.Nonce = ""
	er = txNoNonce.ValidateBasic()
	require.Equal(t, ErrEmptyNonce().Code(), er.Code())

	txInvalidMessageAny := proto.Clone(&tx).(*Transaction)
	txInvalidMessageAny.Msg = nil
	er = txInvalidMessageAny.ValidateBasic()
	require.Equal(t, ErrProtoFromAny(er).Code(), er.Code())

	txEmptySig := proto.Clone(&tx).(*Transaction)
	txEmptySig.Signature = nil
	er = txEmptySig.ValidateBasic()
	require.Equal(t, ErrEmptySignature().Code(), er.Code())

	txEmptyPublicKey := proto.Clone(&tx).(*Transaction)
	txEmptyPublicKey.Signature.PublicKey = nil
	er = txEmptyPublicKey.ValidateBasic()
	require.Equal(t, ErrEmptyPublicKey().Code(), er.Code())

	txInvalidPublicKey := proto.Clone(&tx).(*Transaction)
	txInvalidPublicKey.Signature.PublicKey = []byte("publickey")
	err = txInvalidPublicKey.ValidateBasic()
	require.Equal(t, ErrNewPublicKeyFromBytes(err).Code(), err.Code())

	txInvalidSignature := proto.Clone(&tx).(*Transaction)
	tx.Signature.PublicKey = testingSenderPublicKey.Bytes()
	txInvalidSignature.Signature.Signature = []byte("signature")
	er = txInvalidSignature.ValidateBasic()
	require.Equal(t, ErrSignatureVerificationFailed().Code(), er.Code())
}

func newTestingMsgSend(_ *testing.T) Message {
	return &MessageSend{
		FromAddress: testingSenderAddr,
		ToAddress:   testingToAddr,
		Amount:      defaultAmount,
	}
}

func newUnsignedTestingTransaction(t *testing.T) Transaction {
	msg := newTestingMsgSend(t)

	anyMsg, err := codec.GetCodec().ToAny(msg)
	require.NoError(t, err)

	return Transaction{
		Msg:   anyMsg,
		Nonce: fmt.Sprint(crypto.GetNonce()),
	}
}
