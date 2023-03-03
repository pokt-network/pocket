package types

import (
	"bytes"
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

var _ ITransaction = &Transaction{}

// TxFromBytes unmarshals a proto serialized Transaction to a Transaction protobuf
func TxFromBytes(txProtoBz []byte) (*Transaction, error) {
	tx := &Transaction{}
	if err := codec.GetCodec().Unmarshal(txProtoBz, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// TxHash returns the hash of the proto marshaled transaction
func TxHash(txProtoBz []byte) string {
	return crypto.GetHashStringFromBytes(txProtoBz)
}

// `ITransaction` is an interface that helps capture the functions added to the `Transaction` data structure.
// It is unlikely for there to be multiple implementations of this interface in prod.
type ITransaction interface {
	GetMessage() (proto.Message, error)
	Sign(privateKey crypto.PrivateKey) error
	Hash() (string, error)
	SignableBytes() ([]byte, error)
	Bytes() ([]byte, error)
	Equals(tx2 ITransaction) bool
}

// TODO(#556): Update this function to return pocket specific error codes.
func (tx *Transaction) ValidateBasic() error {
	// Nonce cannot be empty to avoid transaction replays
	if tx.Nonce == "" {
		return fmt.Errorf("nonce cannot be empty") // ErrEmptyNonce
	}

	// Is there a signature we can verify?
	if tx.Signature == nil {
		return fmt.Errorf("signature cannot be empty") // ErrEmptySignature
	}
	if err := tx.Signature.ValidateBasic(); err != nil {
		return err
	}

	// Does the transaction have a valid key?
	publicKey, err := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if err != nil {
		return err // ErrEmptyPublicKey or ErrNewPublicKeyFromBytes
	}

	// Is there a valid msg that can be decoded?
	if _, err := tx.GetMessage(); err != nil {
		return err // ? ErrBadMessage
	}

	signBytes, err := tx.SignableBytes()
	if err != nil {
		return err // ? ErrBadSignature
	}

	if ok := publicKey.Verify(signBytes, tx.Signature.Signature); !ok {
		return fmt.Errorf("signature verification failed") // ErrSignatureVerificationFailed
	}

	return nil
}

func (tx *Transaction) GetMessage() (proto.Message, error) {
	anyMsg, err := codec.GetCodec().FromAny(tx.Msg)
	if err != nil {
		return nil, err
	}
	return anyMsg, nil
}

func (tx *Transaction) Sign(privateKey crypto.PrivateKey) error {
	txSignableBz, err := tx.SignableBytes()
	if err != nil {
		return err
	}
	signature, er := privateKey.Sign(txSignableBz)
	if er != nil {
		return err
	}
	tx.Signature = &Signature{
		PublicKey: privateKey.PublicKey().Bytes(),
		Signature: signature,
	}
	return nil
}

func (tx *Transaction) Hash() (string, error) {
	txProtoBz, err := tx.Bytes()
	if err != nil {
		return "", err
	}
	return TxHash(txProtoBz), nil
}

// The bytes of the transaction that should have been signed.
func (tx *Transaction) SignableBytes() ([]byte, error) {
	// All the contents of the transaction (including the nonce), with the exception of the signature
	// need to be signed by the signer.
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
