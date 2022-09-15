package types

import (
	"bytes"
	"encoding/hex"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

func TransactionFromBytes(transaction []byte) (*Transaction, Error) {
	tx := &Transaction{}
	if err := codec.GetCodec().Unmarshal(transaction, tx); err != nil {
		return nil, ErrUnmarshalTransaction(err)
	}
	return tx, nil
}

func (tx *Transaction) ValidateBasic() Error {
	if tx.Nonce == "" {
		return ErrEmptyNonce()
	}
	if _, err := codec.GetCodec().FromAny(tx.Msg); err != nil {
		return ErrProtoFromAny(err)
	}
	if tx.Signature == nil || tx.Signature.Signature == nil {
		return ErrEmptySignature()
	}
	if tx.Signature.PublicKey == nil {
		return ErrEmptyPublicKey()
	}
	publicKey, err := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if err != nil {
		return ErrNewPublicKeyFromBytes(err)
	}
	signBytes, err := tx.SignBytes()
	if err != nil {
		return ErrProtoMarshal(err)
	}
	if ok := publicKey.Verify(signBytes, tx.Signature.Signature); !ok {
		return ErrSignatureVerificationFailed()
	}
	if _, err := tx.Message(); err != nil {
		return err
	}
	return nil
}

func (tx *Transaction) Message() (Message, Error) {
	codec := codec.GetCodec()
	msg, er := codec.FromAny(tx.Msg)
	if er != nil {
		return nil, ErrProtoMarshal(er)
	}
	message, ok := msg.(Message)
	if !ok {
		return nil, ErrDecodeMessage()
	}
	return message, nil
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) Error {
	publicKey := privateKey.PublicKey()
	bz, err := tx.SignBytes()
	if err != nil {
		return err
	}
	signature, er := privateKey.Sign(bz)
	if er != nil {
		return ErrTransactionSign(er)
	}
	tx.Signature = &Signature{
		PublicKey: publicKey.Bytes(),
		Signature: signature,
	}
	return nil
}

func (tx *Transaction) Hash() (string, Error) {
	b, err := tx.Bytes()
	if err != nil {
		return "", ErrProtoMarshal(err)
	}
	return TransactionHash(b), nil
}

func (tx *Transaction) SignBytes() ([]byte, Error) {
	// transaction := proto.Clone(tx).(*Transaction)
	transaction := *tx
	transaction.Signature = nil
	bz, err := codec.GetCodec().Marshal(&transaction)
	if err != nil {
		return nil, ErrProtoMarshal(err)
	}
	return bz, nil
}

func (tx *Transaction) Bytes() ([]byte, Error) {
	bz, err := codec.GetCodec().Marshal(tx)
	if err != nil {
		return nil, ErrProtoMarshal(err)
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
