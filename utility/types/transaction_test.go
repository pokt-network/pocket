package types

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
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
	if err != nil {
		t.Fatal()
	}
	hash2, err := tx2.Hash()
	if err != nil {
		t.Fatal()
	}
	if hash1 != hash2 {
		t.Fatal("unequal hashes")
	}
	if !tx.Equals(tx2) {
		t.Fatal("unequal transactions")
	}
}

func TestTransaction_Message(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	msg, err := tx.Message()
	require.NoError(t, err)
	expected := NewTestingMsg(t)
	if msg.ProtoReflect().Type() != expected.ProtoReflect().Type() {
		t.Fatal("invalid message type")
	}
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
	if err := tx.Sign(testingSenderPrivateKey); err != nil {
		t.Fatal(err)
	}
	msg, err := tx.SignBytes()
	require.NoError(t, err)
	if !testingSenderPublicKey.Verify(msg, tx.Signature.Signature) {
		t.Fatal("invalid signature error")
	}
}

func TestTransaction_ValidateBasic(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	if err := tx.Sign(testingSenderPrivateKey); err != nil {
		t.Fatal(err)
	}
	if err := tx.ValidateBasic(); err != nil {
		t.Fatal(err)
	}
	txNoNonce := tx
	txNoNonce.Nonce = ""
	er := txNoNonce.ValidateBasic()
	require.Equal(t, types.ErrEmptyNonce(), er.Code())

	txInvalidMessageAny := tx
	txInvalidMessageAny.Msg = nil
	if err := txInvalidMessageAny.ValidateBasic(); err.Code() != types.ErrProtoFromAny(err).Code() {
		t.Fatal(err)
	}
	txEmptySig := tx
	txEmptySig.Signature = nil
	er = txEmptySig.ValidateBasic()
	require.Equal(t, types.ErrEmptySignature(), er.Code())

	txEmptyPublicKey := tx
	txEmptyPublicKey.Signature.PublicKey = nil
	er = txEmptyPublicKey.ValidateBasic()
	require.Equal(t, types.ErrEmptyPublicKey(), er.Code())

	txInvalidPublicKey := tx
	txInvalidPublicKey.Signature.PublicKey = []byte("publickey")
	if err := txInvalidPublicKey.ValidateBasic(); err.Code() != types.ErrNewPublicKeyFromBytes(err).Code() {
		t.Fatal(err)
	}
	txInvalidSignature := tx
	tx.Signature.PublicKey = testingSenderPublicKey.Bytes()
	txInvalidSignature.Signature.Signature = []byte("signature")
	er = txInvalidSignature.ValidateBasic()
	require.Equal(t, types.ErrSignatureVerificationFailed(), er.Code())

}
