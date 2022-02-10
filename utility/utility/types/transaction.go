package types

import (
	"bytes"
	"encoding/hex"
	"pocket/utility/shared/crypto"
	"math/big"
)

func TransactionFromBytes(transaction []byte) (*Transaction, Error) {
	tx := &Transaction{}
	if err := UtilityCodec().Unmarshal(transaction, tx); err != nil {
		return nil, ErrUnmarshalTransaction(err)
	}
	return tx, nil
}

func (x *Transaction) ValidateBasic() Error {
	fee := big.NewInt(0)
	if _, ok := fee.SetString(x.Fee, 10); x.Fee == "" || !ok {
		return ErrNewFeeFromString(x.Fee)
	}
	if x.Nonce == "" {
		return ErrEmptyNonce()
	}
	if _, err := UtilityCodec().FromAny(x.Msg); err != nil {
		return ErrProtoFromAny(err)
	}
	if x.Signature == nil || x.Signature.Signature == nil {
		return ErrEmptySignature()
	}
	if x.Signature.PublicKey == nil {
		return ErrEmptyPublicKey()
	}
	publicKey, err := crypto.NewPublicKeyFromBytes(x.Signature.PublicKey)
	if err != nil {
		return ErrNewPublicKeyFromBytes(err)
	}
	signBytes, err := x.SignBytes()
	if err != nil {
		return ErrProtoMarshal(err)
	}
	if ok := publicKey.VerifyBytes(signBytes, x.Signature.Signature); !ok {
		return ErrSignatureVerificationFailed()
	}
	if _, err := x.Message(); err != nil {
		return err
	}
	return nil
}

func (x *Transaction) Message() (Message, Error) {
	cdc := UtilityCodec()
	msg, er := cdc.FromAny(x.Msg)
	if er != nil {
		return nil, er
	}
	message, ok := msg.(Message)
	if !ok {
		return nil, ErrDecodeMessage()
	}
	return message, nil
}

func (x *Transaction) Sign(privateKey crypto.PrivateKey) Error {
	publicKey := privateKey.PublicKey()
	bz, err := x.SignBytes()
	if err != nil {
		return err
	}
	signature, er := privateKey.Sign(bz)
	if er != nil {
		return ErrTransactionSign(er)
	}
	x.Signature = &Signature{
		PublicKey: publicKey.Bytes(),
		Signature: signature,
	}
	return nil
}

func (x *Transaction) Hash() (string, Error) {
	b, err := x.Bytes()
	if err != nil {
		return "", ErrProtoMarshal(err)
	}
	return TransactionHash(b), nil
}

func (x *Transaction) SignBytes() ([]byte, Error) {
	transaction := *x
	transaction.Signature = nil
	bz, err := UtilityCodec().Marshal(&transaction)
	if err != nil {
		return nil, ErrProtoMarshal(err)
	}
	return bz, nil
}

func (x *Transaction) Bytes() ([]byte, Error) {
	bz, err := UtilityCodec().Marshal(x)
	if err != nil {
		return nil, ErrProtoMarshal(err)
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
