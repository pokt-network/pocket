package types

import (
	"bytes"
	"encoding/hex"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"math/big"
)

func TransactionFromBytes(transaction []byte) (*Transaction, types.Error) {
	tx := &Transaction{}
	if err := UtilityCodec().Unmarshal(transaction, tx); err != nil {
		return nil, types.ErrUnmarshalTransaction(err)
	}
	return tx, nil
}

func (x *Transaction) ValidateBasic() types.Error {
	fee := big.Int{}
	if _, ok := fee.SetString(x.Fee, 10); x.Fee == "" || !ok {
		return types.ErrNewFeeFromString(x.Fee)
	}
	if x.Nonce == "" {
		return types.ErrEmptyNonce()
	}
	if _, err := UtilityCodec().FromAny(x.Msg); err != nil {
		return types.ErrProtoFromAny(err)
	}
	if x.Signature == nil || x.Signature.Signature == nil {
		return types.ErrEmptySignature()
	}
	if x.Signature.PublicKey == nil {
		return types.ErrEmptyPublicKey()
	}
	publicKey, err := crypto.NewPublicKeyFromBytes(x.Signature.PublicKey)
	if err != nil {
		return types.ErrNewPublicKeyFromBytes(err)
	}
	signBytes, err := x.SignBytes()
	if err != nil {
		return types.ErrProtoMarshal(err)
	}
	if ok := publicKey.Verify(signBytes, x.Signature.Signature); !ok {
		return types.ErrSignatureVerificationFailed()
	}
	if _, err := x.Message(); err != nil {
		return err
	}
	return nil
}

func (x *Transaction) Message() (Message, types.Error) {
	codec := UtilityCodec()
	msg, er := codec.FromAny(x.Msg)
	if er != nil {
		return nil, er
	}
	message, ok := msg.(Message)
	if !ok {
		return nil, types.ErrDecodeMessage()
	}
	return message, nil
}

func (x *Transaction) Sign(privateKey crypto.PrivateKey) types.Error {
	publicKey := privateKey.PublicKey()
	bz, err := x.SignBytes()
	if err != nil {
		return err
	}
	signature, er := privateKey.Sign(bz)
	if er != nil {
		return types.ErrTransactionSign(er)
	}
	x.Signature = &Signature{
		PublicKey: publicKey.Bytes(),
		Signature: signature,
	}
	return nil
}

func (x *Transaction) Hash() (string, types.Error) {
	b, err := x.Bytes()
	if err != nil {
		return "", types.ErrProtoMarshal(err)
	}
	return TransactionHash(b), nil
}

func (x *Transaction) SignBytes() ([]byte, types.Error) {
	transaction := *x
	transaction.Signature = nil
	bz, err := UtilityCodec().Marshal(&transaction)
	if err != nil {
		return nil, types.ErrProtoMarshal(err)
	}
	return bz, nil
}

func (x *Transaction) Bytes() ([]byte, types.Error) {
	bz, err := UtilityCodec().Marshal(x)
	if err != nil {
		return nil, types.ErrProtoMarshal(err)
	}
	return bz, nil
}

func (x *Transaction) Equals(tx *Transaction) bool {
	b, _ := tx.Bytes()
	b1, _ := x.Bytes()
	return bytes.Equal(b, b1)
}

func TransactionHash(transactionProtoBytes []byte) string {
	return hex.EncodeToString(crypto.SHA3Hash(transactionProtoBytes))
}
