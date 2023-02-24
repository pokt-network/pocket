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
	testingAmount              = "1000"
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
	require.Error(t, er)
	// require.Equal(t, ErrEmptyNonce().Code(), er.Code())

	txInvalidMessageAny := proto.Clone(&tx).(*Transaction)
	txInvalidMessageAny.Msg = nil
	er = txInvalidMessageAny.ValidateBasic()
	require.Error(t, er)
	// require.Equal(t, ErrProtoFromAny(er).Code(), er.Code())

	txEmptySig := proto.Clone(&tx).(*Transaction)
	txEmptySig.Signature = nil
	er = txEmptySig.ValidateBasic()
	require.Error(t, er)
	// require.Equal(t, ErrEmptySignature().Code(), er.Code())

	txEmptyPublicKey := proto.Clone(&tx).(*Transaction)
	txEmptyPublicKey.Signature.PublicKey = nil
	er = txEmptyPublicKey.ValidateBasic()
	require.Error(t, er)
	// require.Equal(t, ErrEmptyPublicKey().Code(), er.Code())

	txInvalidPublicKey := proto.Clone(&tx).(*Transaction)
	txInvalidPublicKey.Signature.PublicKey = []byte("publickey")
	err = txInvalidPublicKey.ValidateBasic()
	require.Error(t, er)
	// require.Equal(t, ErrNewPublicKeyFromBytes(err).Code(), err.Code())

	txInvalidSignature := proto.Clone(&tx).(*Transaction)
	tx.Signature.PublicKey = testingSenderPublicKey.Bytes()
	txInvalidSignature.Signature.Signature = []byte("signature2")
	er = txInvalidSignature.ValidateBasic()
	require.Error(t, er)
	// require.Equal(t, ErrSignatureVerificationFailed().Code(), er.Code())
}

func newUnsignedTestingTransaction(t *testing.T) Transaction {
	txMsg := &Transaction{}
	anyMsg, err := codec.GetCodec().ToAny(txMsg)
	require.NoError(t, err)

	return Transaction{
		Msg:   anyMsg,
		Nonce: fmt.Sprint(crypto.GetNonce()),
	}
}
