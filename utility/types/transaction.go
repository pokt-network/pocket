package types

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/proto"
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
	msg, er := codec.GetCodec().FromAny(tx.Msg)
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
	sig := tx.Signature // Backup signature
	tx.Signature = nil
	bz, err := codec.GetCodec().Marshal(tx)
	if err != nil {
		return nil, ErrProtoMarshal(err)
	}
	tx.Signature = sig // Restore signature
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

var _ modules.TxResult = &DefaultTxResult{}

func (x *DefaultTxResult) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(x)
}

func (*DefaultTxResult) FromBytes(bz []byte) (modules.TxResult, error) {
	result := new(DefaultTxResult)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (x *DefaultTxResult) Hash() ([]byte, error) {
	bz, err := x.Bytes()
	if err != nil {
		return nil, err
	}
	return x.HashFromBytes(bz)
}

func (x *DefaultTxResult) HashFromBytes(bz []byte) ([]byte, error) {
	return crypto.SHA3Hash(bz), nil
}

func (tx *Transaction) ToTxResult(height int64, index int, signer, recipient, msgType string, er Error) (*DefaultTxResult, Error) {
	txBytes, err := tx.Bytes()
	if err != nil {
		return nil, err
	}
	code, errString := int32(0), ""
	if er != nil {
		code = int32(er.Code())
		errString = err.Error()
	}
	return &DefaultTxResult{
		Tx:            txBytes,
		Height:        height,
		Index:         int32(index),
		ResultCode:    code,
		Error:         errString,
		SignerAddr:    signer,
		RecipientAddr: recipient,
		MessageType:   msgType,
	}, nil
}

func (tx *Transaction) GetMessage() (proto.Message, error) {
	return codec.GetCodec().FromAny(tx.Msg)
}

func TransactionHash(transactionProtoBytes []byte) string {
	return crypto.GetHashStringFromBytes(transactionProtoBytes)
}
