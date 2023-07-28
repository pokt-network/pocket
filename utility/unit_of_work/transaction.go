package unit_of_work

import (
	"bytes"
	"fmt"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// HandleTransaction implements the exposed functionality of the shared utilityUnitOfWork interface.

// IMPROVE: hydration should accept and return the same type (i.e. IndexedTransaction) so there may be opportunity
// to refactor this in the future.
func (u *baseUtilityUnitOfWork) HandleTransaction(tx *coreTypes.Transaction, index int) (*coreTypes.IndexedTransaction, coreTypes.Error) {
	msg, err := u.basicValidateTransaction(tx)
	if err != nil {
		return nil, err
	}
	if u.logger.GetLevel().String() == "debug" {
		u.logger.Debug().Msgf("handling transaction: %+v", msg)
	}
	msgHandlingResult := u.handleMessage(msg)
	return typesUtil.TxToIdxTx(tx, u.height, index, msg, msgHandlingResult)
}

// basicValidateTransaction handles basic transaction validation that is shared across all Messages.
// If basic validation passes (e.g. sufficient fees), the internal message is returned
func (u *baseUtilityUnitOfWork) basicValidateTransaction(tx *coreTypes.Transaction) (typesUtil.Message, coreTypes.Error) {
	// Check if the transaction has a valid message
	msg, err := u.validateTxMessage(tx)
	if err != nil {
		return nil, err
	}

	// Get the address of the transaction signer
	pubKey, er := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if er != nil {
		return nil, coreTypes.ErrNewPublicKeyFromBytes(er)
	}
	address := pubKey.Address()

	// Validate that the signer has a valid signature
	address, err = u.validateTxSignature(address, msg)
	if err != nil {
		return nil, err
	}
	// Update the address of the message signer based on the validated signature
	msg.SetSigner(address)

	// Validate that the signer has enough funds to pay the fee of the message signed
	// and deduct the fee from the signer's account if so.
	if err := u.validateAndDeductTxFees(address, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// validateTxMessage validates the Transaction contains a well-formed messages and returns it if so
func (u *baseUtilityUnitOfWork) validateTxMessage(tx *coreTypes.Transaction) (typesUtil.Message, coreTypes.Error) {
	anyMsg, er := tx.GetMessage()
	if er != nil {
		return nil, coreTypes.ErrDecodeMessage(er)
	}
	msg, ok := anyMsg.(typesUtil.Message)
	if !ok {
		return nil, coreTypes.ErrDecodeMessage(fmt.Errorf("not a supported message type"))
	}
	return msg, nil
}

// validateTxSignature validates that the message has a valid signature from one of the
// candidates and returns the signer's address if so.
func (u *baseUtilityUnitOfWork) validateTxSignature(address crypto.Address, msg typesUtil.Message) ([]byte, coreTypes.Error) {
	signerCandidates, err := u.getSignerCandidates(msg)
	if err != nil {
		return nil, err
	}
	for _, candidate := range signerCandidates {
		if bytes.Equal(candidate, address) {
			return address, nil
		}
	}
	// If no valid signer was found, return an error
	return nil, coreTypes.ErrInvalidSigner(address.ToString())
}

// validateAndDeductTxFees validates that the signer has enough funds to pay the fee of the message signed
// and updates the amounts accordingly if so.
func (u *baseUtilityUnitOfWork) validateAndDeductTxFees(address crypto.Address, msg typesUtil.Message) coreTypes.Error {
	// Retrieve the amounts and fees
	fee, err := u.getFee(msg, msg.GetActorType())
	if err != nil {
		return err
	}
	accountAmount, err := u.getAccountAmount(address)
	if err != nil {
		return coreTypes.ErrGetAccountAmount(err)
	}

	// Validate the account can afford the fees
	accountAmount.Sub(accountAmount, fee)
	if accountAmount.Sign() == -1 {
		return coreTypes.ErrInsufficientAmount(address.String())
	}

	// Remove the fee from the signer's account and add it to the fee collector pool
	if err := u.setAccountAmount(address, accountAmount); err != nil {
		return err
	}
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_FEE_COLLECTOR.Address(), fee); err != nil {
		return err
	}

	return nil
}
