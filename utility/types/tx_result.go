package types

import (
	"encoding/hex"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

func TxToTxResult(
	tx *coreTypes.Transaction,
	height int64,
	index int,
	msg Message,
	msgHandlingResult coreTypes.Error,
) (*coreTypes.TxResult, coreTypes.Error) {
	txBz, err := tx.Bytes()
	if err != nil {
		return nil, coreTypes.ErrProtoMarshal(err)
	}
	resultCode := int32(0)
	errorMsg := ""
	if msgHandlingResult != nil {
		resultCode = int32(msgHandlingResult.Code())
		errorMsg = msgHandlingResult.Error()
	}
	result := &coreTypes.TxResult{
		Tx:            txBz,
		Height:        height,
		Index:         int32(index),
		ResultCode:    resultCode, // TECHDEBT: Remove or update this appropriately.
		Error:         errorMsg,   // TECHDEBT: Remove or update this appropriately.
		SignerAddr:    hex.EncodeToString(msg.GetSigner()),
		RecipientAddr: msg.GetMessageRecipient(),
		MessageType:   msg.GetMessageName(),
	}
	return result, nil
}
