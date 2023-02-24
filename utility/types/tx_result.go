package types

import (
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// INVESTIGATE: Look into a way of removing this type altogether or from shared interfaces.

var _ modules.TxResult = &TxResult{}

func (txr *TxResult) Bytes() ([]byte, error) {
	return codec.GetCodec().Marshal(txr)
}

func (*TxResult) FromBytes(bz []byte) (modules.TxResult, error) {
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

func TxToTxResult(tx *coreTypes.Transaction, height int64, index int, signer, recipient, msgType string, err Error) (*TxResult, Error) {
	txBytes, er := tx.Bytes()
	if er != nil {
		return nil, ErrProtoMarshal(er)
	}
	code, errString := int32(0), ""
	return &TxResult{
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
