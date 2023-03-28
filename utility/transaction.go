package utility

import (
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// HandleTransaction implements the exposed functionality of the shared utilityModule interface.
func (u *utilityModule) HandleTransaction(txProtoBytes []byte) error {
	txHash := coreTypes.TxHash(txProtoBytes)

	// Is the tx already in the mempool (in memory)?
	if u.mempool.Contains(txHash) {
		return typesUtil.ErrDuplicateTransaction()
	}

	// Is the tx already committed & indexed (on disk)?
	if txExists, err := u.GetBus().GetPersistenceModule().TransactionExists(txHash); err != nil {
		return err
	} else if txExists {
		return typesUtil.ErrTransactionAlreadyCommitted()
	}

	// Can the tx be decoded?
	tx := &coreTypes.Transaction{}
	if err := codec.GetCodec().Unmarshal(txProtoBytes, tx); err != nil {
		return typesUtil.ErrProtoUnmarshal(err)
	}

	// Does the tx pass basic validation?
	if err := tx.ValidateBasic(); err != nil {
		return err
	}

	// Store the tx in the mempool
	return u.mempool.AddTx(txProtoBytes)
}
