package types

import (
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

// Bytes serialises TxResult to a byte slice
func (txr *TxResult) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(txr)
}

// FromBytes deserialises a byte slice to a TxResult instance
func (*TxResult) FromBytes(bz []byte) (*TxResult, error) {
	result := new(TxResult)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

// Hash returns the SHA3 hash bytes of the serialised TxResult
func (txr *TxResult) Hash() ([]byte, error) {
	bz, err := txr.Bytes()
	if err != nil {
		return nil, err
	}
	return txr.HashFromBytes(bz), nil
}

// HashFromBytes returns the SHA3 hash bytes of the given byte slice
func (txr *TxResult) HashFromBytes(bz []byte) []byte {
	return crypto.SHA3Hash(bz)
}
