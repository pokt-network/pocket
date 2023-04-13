package types

import (
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
)

func (txr *TxResult) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(txr)
}

func (*TxResult) FromBytes(bz []byte) (*TxResult, error) {
	result := new(TxResult)
	if err := codec.GetCodec().Unmarshal(bz, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (txr *TxResult) Hash() ([]byte, error) {
	bz, err := txr.Bytes()
	if err != nil {
		return nil, err
	}
	return txr.HashFromBytes(bz)
}

func (txr *TxResult) HashFromBytes(bz []byte) ([]byte, error) {
	return crypto.SHA3Hash(bz), nil
}
