package utility

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) ApplyTransaction(tx *typesUtil.Transaction) types.Error {
	msg, err := u.AnteHandleMessage(tx)
	if err != nil {
		return err
	}
	return u.HandleMessage(msg)
}

func (u *UtilityContext) CheckTransaction(transactionProtoBytes []byte) error {
	// validate transaction
	txHash := typesUtil.TransactionHash(transactionProtoBytes)
	if u.Mempool.Contains(txHash) {
		return types.ErrDuplicateTransaction()
	}
	store := u.Store()
	if store.TransactionExists(txHash) { // TODO non-ordered nonce requires non-pruned tx indexer
		return types.ErrTransactionAlreadyCommitted()
	}
	cdc := u.Codec()
	transaction := &typesUtil.Transaction{}
	if err := cdc.Unmarshal(transactionProtoBytes, transaction); err != nil {
		return types.ErrProtoUnmarshal(err)
	}
	if err := transaction.ValidateBasic(); err != nil {
		return err
	}
	// store in mempool
	return u.Mempool.AddTransaction(transactionProtoBytes)
}

func (u *UtilityContext) GetTransactionsForProposal(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) ([][]byte, error) {
	if err := u.BeginBlock(lastBlockByzantineValidators); err != nil {
		return nil, err
	}
	transactions := make([][]byte, 0)
	totalSizeInBytes := 0
	for u.Mempool.Size() != typesUtil.ZeroInt {
		txBytes, err := u.Mempool.PopTransaction()
		if err != nil {
			return nil, err
		}
		transaction, err := typesUtil.TransactionFromBytes(txBytes)
		if err != nil {
			return nil, err
		}
		txSizeInBytes := len(txBytes)
		totalSizeInBytes += txSizeInBytes
		if totalSizeInBytes >= maxTransactionBytes {
			// Add back popped transaction to be applied in a future block
			err := u.Mempool.AddTransaction(txBytes)
			if err != nil {
				return nil, err
			}
			break // we've reached our max
		}
		err = u.ApplyTransaction(transaction)
		if err != nil {
			if err := u.RevertLastSavePoint(); err != nil { // TODO(Andrew): Properly implement 'unhappy path' for save points
				return nil, err
			}
			totalSizeInBytes -= txSizeInBytes
		}
		transactions = append(transactions, txBytes)
	}
	if err := u.EndBlock(proposer); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (u *UtilityContext) AnteHandleMessage(tx *typesUtil.Transaction) (typesUtil.Message, types.Error) {
	msg, err := tx.Message()
	if err != nil {
		return nil, err
	}
	fee, err := u.GetFee(msg) // TODO this enforces exact fee spent regardless of what's put in field... should we remove the fee field from transaction?
	if err != nil {
		return nil, err
	}
	pubKey, er := crypto.NewPublicKeyFromBytes(tx.Signature.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	address := pubKey.Address()
	accountAmount, err := u.GetAccountAmount(address)
	if err != nil {
		return nil, types.ErrGetAccountAmount(err)
	}
	accountAmount.Sub(accountAmount, fee)
	if accountAmount.Sign() == -1 {
		return nil, types.ErrInsufficientAmountError()
	}
	signerCandidates, err := u.GetSignerCandidates(msg)
	if err != nil {
		return nil, err
	}
	var isValidSigner bool
	for _, candidate := range signerCandidates {
		if bytes.Equal(candidate, address) {
			isValidSigner = true
			break
		}
	}
	if !isValidSigner {
		return nil, types.ErrInvalidSigner()
	}
	if err := u.SetAccountAmount(address, accountAmount); err != nil {
		return nil, err
	}
	if err := u.AddPoolAmount(typesUtil.FeePoolName, fee); err != nil {
		return nil, err
	}
	msg.SetSigner(address)
	return msg, nil
}

func (u *UtilityContext) HandleMessage(msg typesUtil.Message) types.Error {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.HandleMessageDoubleSign(x)
	case *typesUtil.MessageSend:
		return u.HandleMessageSend(x)
	case *typesUtil.MessageStakeFisherman:
		return u.HandleMessageStakeFisherman(x)
	case *typesUtil.MessageEditStakeFisherman:
		return u.HandleMessageEditStakeFisherman(x)
	case *typesUtil.MessageUnstakeFisherman:
		return u.HandleMessageUnstakeFisherman(x)
	case *typesUtil.MessagePauseFisherman:
		return u.HandleMessagePauseFisherman(x)
	case *typesUtil.MessageUnpauseFisherman:
		return u.HandleMessageUnpauseFisherman(x)
	case *typesUtil.MessageFishermanPauseServiceNode:
		return u.HandleMessageFishermanPauseServiceNode(x)
	//case *types.MessageTestScore:
	//	return u.HandleMessageTestScore(x)
	//case *types.MessageProveTestScore:
	//	return u.HandleMessageProveTestScore(x)
	case *typesUtil.MessageStakeApp:
		return u.HandleMessageStakeApp(x)
	case *typesUtil.MessageEditStakeApp:
		return u.HandleMessageEditStakeApp(x)
	case *typesUtil.MessageUnstakeApp:
		return u.HandleMessageUnstakeApp(x)
	case *typesUtil.MessagePauseApp:
		return u.HandleMessagePauseApp(x)
	case *typesUtil.MessageUnpauseApp:
		return u.HandleMessageUnpauseApp(x)
	case *typesUtil.MessageStakeValidator:
		return u.HandleMessageStakeValidator(x)
	case *typesUtil.MessageEditStakeValidator:
		return u.HandleMessageEditStakeValidator(x)
	case *typesUtil.MessageUnstakeValidator:
		return u.HandleMessageUnstakeValidator(x)
	case *typesUtil.MessagePauseValidator:
		return u.HandleMessagePauseValidator(x)
	case *typesUtil.MessageUnpauseValidator:
		return u.HandleMessageUnpauseValidator(x)
	case *typesUtil.MessageStakeServiceNode:
		return u.HandleMessageStakeServiceNode(x)
	case *typesUtil.MessageEditStakeServiceNode:
		return u.HandleMessageEditStakeServiceNode(x)
	case *typesUtil.MessageUnstakeServiceNode:
		return u.HandleMessageUnstakeServiceNode(x)
	case *typesUtil.MessagePauseServiceNode:
		return u.HandleMessagePauseServiceNode(x)
	case *typesUtil.MessageUnpauseServiceNode:
		return u.HandleMessageUnpauseServiceNode(x)
	case *typesUtil.MessageChangeParameter:
		return u.HandleMessageChangeParameter(x)
	default:
		return types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetSignerCandidates(msg typesUtil.Message) ([][]byte, types.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignSignerCandidates(x)
	case *typesUtil.MessageSend:
		return u.GetMessageSendSignerCandidates(x)
	case *typesUtil.MessageStakeFisherman:
		return u.GetMessageStakeFishermanSignerCandidates(x)
	case *typesUtil.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanSignerCandidates(x)
	case *typesUtil.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanSignerCandidates(x)
	case *typesUtil.MessagePauseFisherman:
		return u.GetMessagePauseFishermanSignerCandidates(x)
	case *typesUtil.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanSignerCandidates(x)
	case *typesUtil.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeSignerCandidates(x)
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreSignerCandidates(x)
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreSignerCandidates(x)
	case *typesUtil.MessageStakeApp:
		return u.GetMessageStakeAppSignerCandidates(x)
	case *typesUtil.MessageEditStakeApp:
		return u.GetMessageEditStakeAppSignerCandidates(x)
	case *typesUtil.MessageUnstakeApp:
		return u.GetMessageUnstakeAppSignerCandidates(x)
	case *typesUtil.MessagePauseApp:
		return u.GetMessagePauseAppSignerCandidates(x)
	case *typesUtil.MessageUnpauseApp:
		return u.GetMessageUnpauseAppSignerCandidates(x)
	case *typesUtil.MessageStakeValidator:
		return u.GetMessageStakeValidatorSignerCandidates(x)
	case *typesUtil.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorSignerCandidates(x)
	case *typesUtil.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorSignerCandidates(x)
	case *typesUtil.MessagePauseValidator:
		return u.GetMessagePauseValidatorSignerCandidates(x)
	case *typesUtil.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorSignerCandidates(x)
	case *typesUtil.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeSignerCandidates(x)
	case *typesUtil.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeSignerCandidates(x)
	case *typesUtil.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeSignerCandidates(x)
	case *typesUtil.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeSignerCandidates(x)
	case *typesUtil.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeSignerCandidates(x)
	case *typesUtil.MessageChangeParameter:
		return u.GetMessageChangeParameterSignerCandidates(x)
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}
