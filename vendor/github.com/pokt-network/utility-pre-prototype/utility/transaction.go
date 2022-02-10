package utility

import (
	"bytes"
	"github.com/pokt-network/utility-pre-prototype/shared/crypto"
	"github.com/pokt-network/utility-pre-prototype/utility/types"
)

func (u *UtilityContext) ApplyTransaction(tx *types.Transaction) types.Error {
	msg, err := u.AnteHandleMessage(tx)
	if err != nil {
		return err
	}
	return u.HandleMessage(msg)
}

func (u *UtilityContext) CheckTransaction(transactionProtoBytes []byte) error {
	// validate transaction
	txHash := types.TransactionHash(transactionProtoBytes)
	if u.Mempool.Contains(txHash) {
		return types.ErrDuplicateTransaction()
	}
	store := u.Store()
	if store.TransactionExists(txHash) { // TODO non ordered nonce requires non-pruned tx indexer
		return types.ErrTransactionAlreadyCommitted()
	}
	cdc := u.Codec()
	transaction := &types.Transaction{}
	err := cdc.Unmarshal(transactionProtoBytes, transaction)
	if err != nil {
		return types.ErrProtoUnmarshal(err)
	}
	if err := transaction.ValidateBasic(); err != nil {
		return err
	}
	// store in mempool
	return u.Mempool.AddTransaction(transaction)
}

func (u *UtilityContext) GetTransactionsForProposal(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error) {
	if err := u.BeginBlock(lastBlockByzantineValidators); err != nil {
		return nil, err
	}
	totalSizeInBytes := 0
	for u.Mempool.Size() != 0 {
		txBytes, transaction, txSizeInBytes, err := u.Mempool.PopTransaction()
		if err != nil {
			return nil, err
		}
		totalSizeInBytes += txSizeInBytes
		if totalSizeInBytes >= maxTransactionBytes {
			err := u.Mempool.AddTransaction(transaction)
			if err != nil {
				return nil, err
			}
			break // we've reached our max
		}
		err = u.ApplyTransaction(transaction)
		if err != nil {
			if err := u.RevertLastSavePoint(); err != nil {
				return nil, err
			}
			totalSizeInBytes -= txSizeInBytes
		}
		transactions = append(transactions, txBytes)
	}
	if err := u.EndBlock(proposer); err != nil {
		return nil, err
	}
	return
}

func (u *UtilityContext) AnteHandleMessage(tx *types.Transaction) (msg types.Message, err types.Error) {
	msg, err = tx.Message()
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
		return
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
	if err := u.SetAccount(address, accountAmount); err != nil {
		return nil, err
	}
	if err := u.AddPoolAmount(types.FeePoolName, fee); err != nil {
		return nil, err
	}
	msg.SetSigner(address)
	return msg, nil
}

func (u *UtilityContext) HandleMessage(msg types.Message) types.Error {
	switch x := msg.(type) {
	case *types.MessageDoubleSign:
		return u.HandleMessageDoubleSign(x)
	case *types.MessageSend:
		return u.HandleMessageSend(x)
	case *types.MessageStakeFisherman:
		return u.HandleMessageStakeFisherman(x)
	case *types.MessageEditStakeFisherman:
		return u.HandleMessageEditStakeFisherman(x)
	case *types.MessageUnstakeFisherman:
		return u.HandleMessageUnstakeFisherman(x)
	case *types.MessagePauseFisherman:
		return u.HandleMessagePauseFisherman(x)
	case *types.MessageUnpauseFisherman:
		return u.HandleMessageUnpauseFisherman(x)
	case *types.MessageFishermanPauseServiceNode:
		return u.HandleMessageFishermanPauseServiceNode(x)
	//case *types.MessageTestScore:
	//	return u.HandleMessageTestScore(x)
	//case *types.MessageProveTestScore:
	//	return u.HandleMessageProveTestScore(x)
	case *types.MessageStakeApp:
		return u.HandleMessageStakeApp(x)
	case *types.MessageEditStakeApp:
		return u.HandleMessageEditStakeApp(x)
	case *types.MessageUnstakeApp:
		return u.HandleMessageUnstakeApp(x)
	case *types.MessagePauseApp:
		return u.HandleMessagePauseApp(x)
	case *types.MessageUnpauseApp:
		return u.HandleMessageUnpauseApp(x)
	case *types.MessageStakeValidator:
		return u.HandleMessageStakeValidator(x)
	case *types.MessageEditStakeValidator:
		return u.HandleMessageEditStakeValidator(x)
	case *types.MessageUnstakeValidator:
		return u.HandleMessageUnstakeValidator(x)
	case *types.MessagePauseValidator:
		return u.HandleMessagePauseValidator(x)
	case *types.MessageUnpauseValidator:
		return u.HandleMessageUnpauseValidator(x)
	case *types.MessageStakeServiceNode:
		return u.HandleMessageStakeServiceNode(x)
	case *types.MessageEditStakeServiceNode:
		return u.HandleMessageEditStakeServiceNode(x)
	case *types.MessageUnstakeServiceNode:
		return u.HandleMessageUnstakeServiceNode(x)
	case *types.MessagePauseServiceNode:
		return u.HandleMessagePauseServiceNode(x)
	case *types.MessageUnpauseServiceNode:
		return u.HandleMessageUnpauseServiceNode(x)
	case *types.MessageChangeParameter:
		return u.HandleMessageChangeParameter(x)
	default:
		return types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetSignerCandidates(msg types.Message) (candidates [][]byte, err types.Error) {
	switch x := msg.(type) {
	case *types.MessageDoubleSign:
		return u.GetMessageDoubleSignSignerCandidates(x)
	case *types.MessageSend:
		return u.GetMessageSendSignerCandidates(x)
	case *types.MessageStakeFisherman:
		return u.GetMessageStakeFishermanSignerCandidates(x)
	case *types.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanSignerCandidates(x)
	case *types.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanSignerCandidates(x)
	case *types.MessagePauseFisherman:
		return u.GetMessagePauseFishermanSignerCandidates(x)
	case *types.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanSignerCandidates(x)
	case *types.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeSignerCandidates(x)
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreSignerCandidates(x)
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreSignerCandidates(x)
	case *types.MessageStakeApp:
		return u.GetMessageStakeAppSignerCandidates(x)
	case *types.MessageEditStakeApp:
		return u.GetMessageEditStakeAppSignerCandidates(x)
	case *types.MessageUnstakeApp:
		return u.GetMessageUnstakeAppSignerCandidates(x)
	case *types.MessagePauseApp:
		return u.GetMessagePauseAppSignerCandidates(x)
	case *types.MessageUnpauseApp:
		return u.GetMessageUnpauseAppSignerCandidates(x)
	case *types.MessageStakeValidator:
		return u.GetMessageStakeValidatorSignerCandidates(x)
	case *types.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorSignerCandidates(x)
	case *types.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorSignerCandidates(x)
	case *types.MessagePauseValidator:
		return u.GetMessagePauseValidatorSignerCandidates(x)
	case *types.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorSignerCandidates(x)
	case *types.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeSignerCandidates(x)
	case *types.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeSignerCandidates(x)
	case *types.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeSignerCandidates(x)
	case *types.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeSignerCandidates(x)
	case *types.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeSignerCandidates(x)
	case *types.MessageChangeParameter:
		return u.GetMessageChangeParameterSignerCandidates(x)
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}
