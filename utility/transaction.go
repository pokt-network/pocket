package utility

import (
	"bytes"
	"encoding/hex"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *utilityModule) HandleTransaction(txProtoBytes []byte) error {
	// Is the tx already in the mempool (in memory)?
	txHash := coreTypes.TxHash(txProtoBytes)
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

func (u *utilityContext) applyTx(index int, tx *coreTypes.Transaction) (modules.TxResult, typesUtil.Error) {
	msg, signer, err := u.anteHandleMessage(tx)
	if err != nil {
		return nil, err
	}
	return typesUtil.TxToTxResult(tx, u.height, index, signer, msg.GetMessageRecipient(), msg.GetMessageName(), u.handleMessage(msg))
}

func (u *utilityContext) anteHandleMessage(tx *coreTypes.Transaction) (msg typesUtil.Message, signer string, err typesUtil.Error) {
	anyMsg, er := tx.GetMessage()
	if err != nil {
		return nil, "", typesUtil.ErrDecodeMessage()
	}
	msg, ok := anyMsg.(typesUtil.Message)
	if !ok {
		return nil, "", typesUtil.ErrDecodeMessage()
	}

	fee, err := u.getFee(msg, msg.GetActorType())
	if err != nil {
		return nil, "", err
	}
	pubKey, er := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if er != nil {
		return nil, "", typesUtil.ErrNewPublicKeyFromBytes(er)
	}
	address := pubKey.Address()
	accountAmount, err := u.getAccountAmount(address)
	if err != nil {
		return nil, "", typesUtil.ErrGetAccountAmount(err)
	}
	accountAmount.Sub(accountAmount, fee)
	if accountAmount.Sign() == -1 {
		return nil, "", typesUtil.ErrInsufficientAmount(address.String())
	}
	signerCandidates, err := u.getSignerCandidates(msg)
	if err != nil {
		return nil, "", err
	}
	var isValidSigner bool
	for _, candidate := range signerCandidates {
		if bytes.Equal(candidate, address) {
			isValidSigner = true
			signer = hex.EncodeToString(candidate)
			break
		}
	}
	if !isValidSigner {
		return nil, signer, typesUtil.ErrInvalidSigner()
	}
	if err := u.setAccountAmount(address, accountAmount); err != nil {
		return nil, signer, err
	}
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_FEE_COLLECTOR.FriendlyName(), fee); err != nil {
		return nil, "", err
	}
	msg.SetSigner(address)
	return msg, signer, nil
}
