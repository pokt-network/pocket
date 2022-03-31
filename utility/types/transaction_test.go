package types

import (
	"bytes"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"testing"
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
	codec := UtilityCodec()
	msg := NewTestingMsg(t)
	anyMsg, err := codec.ToAny(msg)
	if err != nil {
		t.Fatal(err)
	}
	return Transaction{
		Msg:   anyMsg,
		Fee:   defaultFee,
		Nonce: types.BigIntToString(types.RandBigInt()),
	}
}

func TestTransactionBytesAndFromBytes(t *testing.T) {
	tx := NewUnsignedTestingTransaction(t)
	bz, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tx2, err := TransactionFromBytes(bz)
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
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
	if err := txNoNonce.ValidateBasic(); err.Code() != types.ErrEmptyNonce().Code() {
		t.Fatal(err)
	}
	txInvalidMessageAny := tx
	txInvalidMessageAny.Msg = nil
	if err := txInvalidMessageAny.ValidateBasic(); err.Code() != types.ErrProtoFromAny(err).Code() {
		t.Fatal(err)
	}
	txEmptySig := tx
	txEmptySig.Signature = nil
	if err := txEmptySig.ValidateBasic(); err.Code() != types.ErrEmptySignature().Code() {
		t.Fatal(err)
	}
	txEmptyPublicKey := tx
	txEmptyPublicKey.Signature.PublicKey = nil
	if err := txEmptyPublicKey.ValidateBasic(); err.Code() != types.ErrEmptyPublicKey().Code() {
		t.Fatal(err)
	}
	txInvalidPublicKey := tx
	txInvalidPublicKey.Signature.PublicKey = []byte("publickey")
	if err := txInvalidPublicKey.ValidateBasic(); err.Code() != types.ErrNewPublicKeyFromBytes(err).Code() {
		t.Fatal(err)
	}
	txInvalidSignature := tx
	tx.Signature.PublicKey = testingSenderPublicKey.Bytes()
	txInvalidSignature.Signature.Signature = []byte("signature")
	if err := txInvalidSignature.ValidateBasic(); err.Code() != types.ErrSignatureVerificationFailed().Code() {
		t.Fatal(err)
	}
}
