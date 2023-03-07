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

	msg, err := tx.SignableBytes()
	require.NoError(t, err)

	verified := testingSenderPublicKey.Verify(msg, tx.Signature.Signature)
	require.True(t, verified, "signature should be verified")
}

func TestTransaction_ValidateBasic(t *testing.T) {
	tx := newUnsignedTestingTransaction(t)
	err := tx.Sign(testingSenderPrivateKey)
	require.NoError(t, err)

	err = tx.ValidateBasic()
	require.NoError(t, err)

	txNoNonce := proto.Clone(&tx).(*Transaction)
	txNoNonce.Nonce = ""
	err = txNoNonce.ValidateBasic()
	require.Error(t, err)

	txInvalidMessageAny := proto.Clone(&tx).(*Transaction)
	txInvalidMessageAny.Msg = nil
	err = txInvalidMessageAny.ValidateBasic()
	require.Error(t, err)

	txEmptySig := proto.Clone(&tx).(*Transaction)
	txEmptySig.Signature = nil
	err = txEmptySig.ValidateBasic()
	require.Error(t, err)

	txEmptyPublicKey := proto.Clone(&tx).(*Transaction)
	txEmptyPublicKey.Signature.PublicKey = nil
	err = txEmptyPublicKey.ValidateBasic()
	require.Error(t, err)

	txInvalidPublicKey := proto.Clone(&tx).(*Transaction)
	txInvalidPublicKey.Signature.PublicKey = []byte("publickey")
	err = txInvalidPublicKey.ValidateBasic()
	require.Error(t, err)

	txInvalidSignature := proto.Clone(&tx).(*Transaction)
	tx.Signature.PublicKey = testingSenderPublicKey.Bytes()
	txInvalidSignature.Signature.Signature = []byte("signature2")
	err = txInvalidSignature.ValidateBasic()
	require.Error(t, err)
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
