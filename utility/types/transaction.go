package types

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

// No need for a Signature interface abstraction for the time being.
// DISCUSS_IN_THIS_COMMIT: Should we create an interface for `Transaction` to capture the other functions it has?
var _ Validatable = &Transaction{}

func TransactionFromBytes(txProtoBytes []byte) (*Transaction, Error) {
	tx := &Transaction{}
	if err := codec.GetCodec().Unmarshal(txProtoBytes, tx); err != nil {
		return nil, ErrUnmarshalTransaction(err)
	}
	return tx, nil
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

func TxHash(txProtoBytes []byte) string {
	return crypto.GetHashStringFromBytes(txProtoBytes)
}

// The bytes of the transaction that should have been signed.
// INVESTIGATE: Should this potentially be `tx.Message().GetCanonicalBytes()` instead?
func (tx *Transaction) SignableBytes() ([]byte, error) {
	transaction := codec.GetCodec().Clone(tx).(*Transaction)
	transaction.Signature = nil
	return codec.GetCodec().Marshal(transaction)
}

// func (tx *Transaction) SignBytes() ([]byte, Error) {
// 	sig := tx.Signature // Backup signature
// 	tx.Signature = nil
// 	bz, err := codec.GetCodec().Marshal(tx)
// 	if err != nil {
// 		return nil, ErrProtoMarshal(err)
// 	}
// 	tx.Signature = sig // Restore signature
// 	return bz, nil

func (tx *Transaction) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(tx)
}

func (tx *Transaction) Equals(tx2 *Transaction) bool {
	b, err := tx.Bytes()
	b2, err2 := tx2.Bytes()
	return err != nil && err2 != nil && bytes.Equal(b, b2)
}
