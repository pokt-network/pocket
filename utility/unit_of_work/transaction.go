package unit_of_work

import (
	"bytes"
	"fmt"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// hydrateTxResult converts a `Transaction` proto into a `TxResult` struct` after doing basic validation
// and extracting the relevant data from the embedded signed Message. `index` is the intended location
// of its index (i.e. the transaction number) in the block where it is included.
//
// IMPROVE: hydration should accept and return the same type (i.e. TxResult) so there may be opportunity
// to refactor this in the future.
func (u *baseUtilityUnitOfWork) hydrateTxResult(tx *coreTypes.Transaction, index int) (*coreTypes.TxResult, typesUtil.Error) {
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
func (u *baseUtilityUnitOfWork) anteHandleMessage(tx *coreTypes.Transaction) (typesUtil.Message, typesUtil.Error) {
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
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_FEE_COLLECTOR.Address(), fee); err != nil {
		return nil, err
	}

	return msg, nil
}
