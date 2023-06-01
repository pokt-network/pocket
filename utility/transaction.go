package utility

import (
	"encoding/hex"
	"errors"

	"github.com/dgraph-io/badger/v3"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// HandleTransaction implements the exposed functionality of the shared utilityModule interface.
func (u *utilityModule) HandleTransaction(txProtoBytes []byte) error {
	txHash := coreTypes.TxHash(txProtoBytes)

	// Is the tx already in the mempool (in memory)?
	if u.mempool.Contains(txHash) {
		return coreTypes.ErrDuplicateTransaction()
	}

	// Is the tx already committed & indexed (on disk)?
	if txExists, err := u.GetBus().GetPersistenceModule().TransactionExists(txHash); err != nil {
		return err
	} else if txExists {
		return coreTypes.ErrTransactionAlreadyCommitted()
	}

	// Can the tx be decoded?
	tx := &coreTypes.Transaction{}
	if err := codec.GetCodec().Unmarshal(txProtoBytes, tx); err != nil {
		return coreTypes.ErrProtoUnmarshal(err)
	}

	// Does the tx pass basic validation?
	if err := tx.ValidateBasic(); err != nil {
		return err
	}

	// Store the tx in the mempool
	return u.mempool.AddTx(txProtoBytes)
}

// GetIndexedTransaction implements the exposed functionality of the shared utilityModule interface.
func (u *utilityModule) GetIndexedTransaction(txProtoBytes []byte) (*coreTypes.IndexedTransaction, error) {
	txHash := coreTypes.TxHash(txProtoBytes)

	// TECHDEBT: Note the inconsistency between referencing tx hash as a string vs. byte slice in different places. Need to pick
	// one and consolidate throughout the codebase
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}
	idTx, err := u.GetBus().GetPersistenceModule().GetTxIndexer().GetByHash(hash)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, coreTypes.ErrTransactionNotCommitted()
		}
		return nil, err
	}

	return idTx, nil
}
