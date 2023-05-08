package types

import (
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Bytes serialises IndexedTransaction to a byte slice
func (tx *IndexedTransaction) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(tx)
}

// FromBytes deserialises a byte slice to a IndexedTransaction instance
func (*IndexedTransaction) FromBytes(bz []byte) (*IndexedTransaction, error) {
	result := new(IndexedTransaction)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

// Hash returns the SHA3 hash bytes of the serialised IndexedTransaction
// CONSIDER: Making Hash use by default the tx.GetTx() to get the Hash
func (tx *IndexedTransaction) Hash() ([]byte, error) {
	bz, err := tx.Bytes()
	if err != nil {
		return nil, err
	}
	return tx.HashFromBytes(bz), nil
}

// HashFromBytes returns the SHA3 hash bytes of the given byte slice
func (tx *IndexedTransaction) HashFromBytes(bz []byte) []byte {
	return crypto.SHA3Hash(bz)
}
