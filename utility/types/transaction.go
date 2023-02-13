package types

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

func TxFromBytes(txProtoBytes []byte) (*Transaction, Error) {
	tx := &Transaction{}
	if err := codec.GetCodec().Unmarshal(txProtoBytes, tx); err != nil {
		return nil, ErrUnmarshalTransaction(err)
	}
	return tx, nil
}

func TxHash(txProtoBytes []byte) string {
	return crypto.GetHashStringFromBytes(txProtoBytes)
}

var (
	_ Validatable  = &Transaction{}
	_ ITransaction = &Transaction{}
)

// `ITransaction` is an interface that helps capture the functions added to the `Transaction` data structure.
// It is unlikely for there to be multiple implementations of this interface in prod.
type ITransaction interface {
	GetMessage() (Message, Error)
	Sign(privateKey crypto.PrivateKey) Error
	Hash() (string, Error)
	SignableBytes() ([]byte, error)
	Bytes() ([]byte, error)
	Equals(tx2 ITransaction) bool
}

func (tx *Transaction) ValidateBasic() Error {
	// Nonce cannot be empty to avoid transaction replays
	if tx.Nonce == "" {
		return ErrEmptyNonce()
	}

	// Is there a signature we can verify?
	if tx.Signature == nil {
		return ErrEmptySignature()
	}
	if err := tx.Signature.ValidateBasic(); err != nil {
		return err
	}

	// Does the transaction have a valid key?
	publicKey, err := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if err != nil {
		return ErrNewPublicKeyFromBytes(err)
	}

	// Is there a valid msg that can be decoded?
	if _, err := tx.GetMessage(); err != nil {
		return err
	}

	signBytes, err := tx.SignableBytes()
	if err != nil {
		return ErrProtoMarshal(err)
	}
	if ok := publicKey.Verify(signBytes, tx.Signature.Signature); !ok {
		return ErrSignatureVerificationFailed()
	}

	return nil
}

func (tx *Transaction) GetMessage() (Message, Error) {
	msg, err := codec.GetCodec().FromAny(tx.Msg)
	if err != nil {
		return nil, ErrProtoFromAny(err)
	}
	message, ok := msg.(Message)
	if !ok {
		return nil, ErrDecodeMessage()
	}
	return message, nil
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) Error {
	txSignableBz, err := tx.SignableBytes()
	if err != nil {
		return ErrProtoMarshal(err)
	}
	signature, er := privateKey.Sign(txSignableBz)
	if er != nil {
		return ErrTransactionSign(er)
	}
	tx.Signature = &Signature{
		PublicKey: privateKey.PublicKey().Bytes(),
		Signature: signature,
	}
	return nil
}

func (tx *Transaction) Hash() (string, Error) {
	txProtoBz, err := tx.Bytes()
	if err != nil {
		return "", ErrProtoMarshal(err)
	}
	return TxHash(txProtoBz), nil
}

// The bytes of the transaction that should have been signed.
func (tx *Transaction) SignableBytes() ([]byte, error) {
	// This is not simply `tx.Message().GetCanonicalBytes()` because the txCopy also contains
	// other metadata such as the nonce which has to be part signed as well.
	txCopy := codec.GetCodec().Clone(tx).(*Transaction)
	txCopy.Signature = nil
	return codec.GetCodec().Marshal(txCopy)
}

func (tx *Transaction) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(tx)
}

func (tx *Transaction) Equals(tx2 ITransaction) bool {
	b, err := tx.Bytes()
	b2, err2 := tx2.Bytes()
	return err != nil && err2 != nil && bytes.Equal(b, b2)
}
