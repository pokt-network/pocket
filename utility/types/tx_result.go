package types

import (
	"encoding/hex"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.TxResult = &TxResult{}

func TxToTxResult(
	tx *coreTypes.Transaction,
	height int64,
	index int,
	msg Message,
	msgHandlingResult Error,
) (*TxResult, Error) {
	txBz, err := tx.Bytes()
	if err != nil {
		return nil, ErrProtoMarshal(err)
	}
	return &TxResult{
		Tx:            txBz,
		Height:        height,
		Index:         int32(index),
		ResultCode:    int32(msgHandlingResult.Code()), // TECHDEBT: Remove or update this appropriately.
		Error:         msgHandlingResult.Error(),       // TECHDEBT: Remove or update this appropriately.
		SignerAddr:    hex.EncodeToString(msg.GetSigner()),
		RecipientAddr: msg.GetMessageRecipient(),
		MessageType:   msg.GetMessageName(),
	}, nil
}

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
