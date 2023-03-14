package utility

import (
	"bytes"
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/pokterrors"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// HandleTransaction implements the exposed functionality of the shared utilityModule interface.
func (u *utilityModule) HandleTransaction(txProtoBytes []byte) error {
	txHash := coreTypes.TxHash(txProtoBytes)

	// Is the tx already in the mempool (in memory)?
	if u.mempool.Contains(txHash) {
		return pokterrors.UtilityErrDuplicateTransaction()
	}

	// Is the tx already committed & indexed (on disk)?
	if txExists, err := u.GetBus().GetPersistenceModule().TransactionExists(txHash); err != nil {
		return err
	} else if txExists {
		return pokterrors.UtilityErrTransactionAlreadyCommitted()
	}

	// Can the tx be decoded?
	tx := &coreTypes.Transaction{}
	if err := codec.GetCodec().Unmarshal(txProtoBytes, tx); err != nil {
		return pokterrors.UtilityErrProtoUnmarshal(err)
	}

	// Does the tx pass basic validation?
	if err := tx.ValidateBasic(); err != nil {
		return err
	}

	// Store the tx in the mempool
	return u.mempool.AddTx(txProtoBytes)
}

// hydrateTxResult converts a `Transaction` proto into a `TxResult` struct` after doing basic validation
// and extracting the relevant data from the embedded signed Message. `index` is the intended location
// of its index (i.e. the transaction number) in the block where it is included.
//
// IMPROVE: hydration should accept and return the same type (i.e. TxResult) so there may be opportunity
// to refactor this in the future.
func (u *utilityContext) hydrateTxResult(tx *coreTypes.Transaction, index int) (modules.TxResult, pokterrors.Error) {
	msg, err := u.anteHandleMessage(tx)
	if err != nil {
		return nil, err
	}
	msgHandlingResult := u.handleMessage(msg)
	return typesUtil.TxToTxResult(tx, u.height, index, msg, msgHandlingResult)
}

// anteHandleMessage handles basic validation of the message in the Transaction before it is processed
// REFACTOR: Splitting this into a `feeValidation`, `signerValidation`, and `messageValidation` etc
// would make it more modular and readable.
func (u *utilityContext) anteHandleMessage(tx *coreTypes.Transaction) (typesUtil.Message, pokterrors.Error) {
	// Check if the transaction has a valid message
	anyMsg, er := tx.GetMessage()
	if er != nil {
		return nil, pokterrors.UtilityErrDecodeMessage(er)
	}
	msg, ok := anyMsg.(typesUtil.Message)
	if !ok {
		return nil, pokterrors.UtilityErrDecodeMessage(fmt.Errorf("not a supported message type"))
	}

	// Get the address of the transaction signer
	pubKey, err := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if err != nil {
		return nil, pokterrors.UtilityErrNewPublicKeyFromBytes(err)
	}
	address := pubKey.Address()
	addressHex := address.ToString()

	// Validate that the signer has enough funds to pay the fee of the message signed
	fee, er := u.getFee(msg, msg.GetActorType())
	if er != nil {
		return nil, er
	}
	accountAmount, err := u.getAccountAmount(address)
	if err != nil {
		return nil, pokterrors.UtilityErrGetAccountAmount(err)
	}
	accountAmount.Sub(accountAmount, fee)
	if accountAmount.Sign() == -1 {
		return nil, pokterrors.UtilityErrInsufficientAmount(addressHex)
	}

	// Validate that the signer has a valid signature
	var isValidSigner bool
	signerCandidates, er := u.getSignerCandidates(msg)
	if err != nil {
		return nil, er
	}
	for _, candidate := range signerCandidates {
		if bytes.Equal(candidate, address) {
			isValidSigner = true
			msg.SetSigner(address)
			break
		}
	}
	if !isValidSigner {
		return nil, pokterrors.UtilityErrInvalidSigner(addressHex)
	}

	// Remove the fee from the signer's account and add it to the fee collector pool
	if err := u.setAccountAmount(address, accountAmount); err != nil {
		return nil, err
	}
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_FEE_COLLECTOR.FriendlyName(), fee); err != nil {
		return nil, err
	}

	return msg, nil
}
