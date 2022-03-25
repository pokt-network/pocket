package types

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
)

func TransactionFromBytes(transaction []byte) (*Transaction, types.Error) {
	tx := &Transaction{}
	if err := UtilityCodec().Unmarshal(transaction, tx); err != nil {
		return nil, types.ErrUnmarshalTransaction(err)
	}
	return tx, nil
}

func (tx *Transaction) ValidateBasic() types.Error {
	fee := big.Int{}
	if _, ok := fee.SetString(tx.Fee, 10); tx.Fee == "" || !ok {
		return types.ErrNewFeeFromString(tx.Fee)
	}
	if tx.Nonce == "" {
		return types.ErrEmptyNonce()
	}
	if _, err := UtilityCodec().FromAny(tx.Msg); err != nil {
		return types.ErrProtoFromAny(err)
	}
	if tx.Signature == nil || tx.Signature.Signature == nil {
		return types.ErrEmptySignature()
	}
	if tx.Signature.PublicKey == nil {
		return types.ErrEmptyPublicKey()
	}
	publicKey, err := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if err != nil {
		return types.ErrNewPublicKeyFromBytes(err)
	}
	signBytes, err := tx.SignBytes()
	if err != nil {
		return types.ErrProtoMarshal(err)
	}
	if ok := publicKey.Verify(signBytes, tx.Signature.Signature); !ok {
		return types.ErrSignatureVerificationFailed()
	}
	if _, err := tx.Message(); err != nil {
		return err
	}
	return nil
}

func (tx *Transaction) Message() (Message, types.Error) {
	codec := UtilityCodec()
	msg, er := codec.FromAny(tx.Msg)
	if er != nil {
		return nil, er
	}
	message, ok := msg.(Message)
	if !ok {
		return nil, types.ErrDecodeMessage()
	}
	return message, nil
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) types.Error {
	publicKey := privateKey.PublicKey()
	bz, err := tx.SignBytes()
	if err != nil {
		return err
	}
	signature, er := privateKey.Sign(bz)
	if er != nil {
		return types.ErrTransactionSign(er)
	}
	tx.Signature = &Signature{
		PublicKey: publicKey.Bytes(),
		Signature: signature,
	}
	return nil
}

func (tx *Transaction) Hash() (string, types.Error) {
	b, err := tx.Bytes()
	if err != nil {
		return "", types.ErrProtoMarshal(err)
	}
	return TransactionHash(b), nil
}

func (tx *Transaction) SignBytes() ([]byte, types.Error) {
	// transaction := proto.Clone(tx).(*Transaction)
	transaction := *tx
	transaction.Signature = nil
	bz, err := UtilityCodec().Marshal(&transaction)
	if err != nil {
		return nil, types.ErrProtoMarshal(err)
	}
	return bz, nil
}

func (tx *Transaction) Bytes() ([]byte, types.Error) {
	bz, err := UtilityCodec().Marshal(tx)
	if err != nil {
		return nil, types.ErrProtoMarshal(err)
	}
	return bz, nil
}

func (tx *Transaction) Equals(tx2 *Transaction) bool {
	b, _ := tx2.Bytes()
	b1, _ := tx2.Bytes()
	return bytes.Equal(b, b1)
}

func TransactionHash(transactionProtoBytes []byte) string {
	return hex.EncodeToString(crypto.SHA3Hash(transactionProtoBytes))
}
