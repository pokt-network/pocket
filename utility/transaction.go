package utility

import (
	"bytes"
	"fmt"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
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

// hydrateTx converts a Transaction into a TxResult after doing basic validation and extracting
// the relevant data from the embedded signed Message.
func (u *utilityContext) hydrateTx(tx *coreTypes.Transaction, index int) (modules.TxResult, typesUtil.Error) {
	msg, err := u.anteHandleMessage(tx)
	if err != nil {
		return nil, err
	}
	msgHandlingResult := u.handleMessage(msg)
	// INCOMPLETE:
	return typesUtil.TxToTxResult(tx, u.height, index, msg, msgHandlingResult)
}

// anteHandleMessage handles basic validation of the message in the Transaction before it is processed
// REFACTOR: Splitting this into a `feeValidation`, `signerValidation`, and `messageValidation` etc
// would make it more modular and readable.
func (u *utilityContext) anteHandleMessage(tx *coreTypes.Transaction) (typesUtil.Message, typesUtil.Error) {
	// Check if the transaction has a valid message
	anyMsg, er := tx.GetMessage()
	if er != nil {
		return nil, typesUtil.ErrDecodeMessage(er)
	}
	msg, ok := anyMsg.(typesUtil.Message)
	if !ok {
		return nil, typesUtil.ErrDecodeMessage(fmt.Errorf("not a supported message type"))
	}

	// Get the address of the transaction signer
	pubKey, er := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if er != nil {
		return nil, typesUtil.ErrNewPublicKeyFromBytes(er)
	}
	address := pubKey.Address()
	addressHex := address.ToString()

	// Validate that the signer has enough funds to pay the fee of the message signed
	fee, err := u.getFee(msg, msg.GetActorType())
	if err != nil {
		return nil, err
	}
	accountAmount, err := u.getAccountAmount(address)
	if err != nil {
		return nil, typesUtil.ErrGetAccountAmount(err)
	}
	accountAmount.Sub(accountAmount, fee)
	if accountAmount.Sign() == -1 {
		return nil, typesUtil.ErrInsufficientAmount(addressHex)
	}

	// Validate that the signer has a valid signature
	var isValidSigner bool
	signerCandidates, err := u.getSignerCandidates(msg)
	if err != nil {
		return nil, err
	}
	for _, candidate := range signerCandidates {
		if bytes.Equal(candidate, address) {
			isValidSigner = true
			msg.SetSigner(address)
			break
		}
	}
	if !isValidSigner {
		return nil, typesUtil.ErrInvalidSigner(addressHex)
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
