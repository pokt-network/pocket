package types

import (
	"encoding/hex"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// TxToIdxTx Converts a Transaction structure into an IndexedTransaction structure
//
// RESEARCH: The reader may notice that even invalid messages result in indexed transaction.
// This is common in other PoS networks, such as Tendermint, so fees can be deducted even if its invalid.
// This is a design decision that will need to be revisited and can have lots repercussions, sybil attack,
// vectors, etc if not properly thought out.
func TxToIdxTx(
	tx *coreTypes.Transaction,
	height int64,
	index int,
	msg Message,
	msgHandlingResult coreTypes.Error,
) (*coreTypes.IndexedTransaction, coreTypes.Error) {
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
	result := &coreTypes.IndexedTransaction{
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
