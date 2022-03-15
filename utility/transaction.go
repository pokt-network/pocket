package utility

import (
	"bytes"
	"fmt"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	types2 "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) ApplyTransaction(tx *types2.Transaction) types.Error {
	msg, err := u.AnteHandleMessage(tx)
	if err != nil {
		return err
	}
	return u.HandleMessage(msg)
}

func (u *UtilityContext) CheckTransaction(transactionProtoBytes []byte) error {
	// validate transaction
	txHash := types2.TransactionHash(transactionProtoBytes)
	if u.Mempool.Contains(txHash) {
		return types.ErrDuplicateTransaction()
	}
	store := u.Store()
	if store.TransactionExists(txHash) { // TODO non ordered nonce requires non-pruned tx indexer
		return types.ErrTransactionAlreadyCommitted()
	}
	cdc := u.Codec()
	transaction := &types2.Transaction{}
	err := cdc.Unmarshal(transactionProtoBytes, transaction)
	if err != nil {
		return types.ErrProtoUnmarshal(err)
	}
	if err := transaction.ValidateBasic(); err != nil {
		return err
	}
	// store in mempool
	return u.Mempool.AddTransaction(transactionProtoBytes)
}

func (u *UtilityContext) GetTransactionsForProposal(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error) {
	if err := u.BeginBlock(lastBlockByzantineValidators); err != nil {
		return nil, err
	}
	totalSizeInBytes := 0
	for u.Mempool.Size() != 0 {
		txBytes, txSizeInBytes, err := u.Mempool.PopTransaction()
		if err != nil {
			return nil, err
		}
		transaction, err := types2.TransactionFromBytes(txBytes)
		if err != nil {
			return nil, err
		}
		totalSizeInBytes += txSizeInBytes
		if totalSizeInBytes >= maxTransactionBytes {
			err := u.Mempool.AddTransaction(txBytes)
			if err != nil {
				return nil, err
			}
			break // we've reached our max
		}
		err = u.ApplyTransaction(transaction)
		if err != nil {
			fmt.Println(err)
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

func (u *UtilityContext) AnteHandleMessage(tx *types2.Transaction) (msg types2.Message, err types.Error) {
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
	if err := u.AddPoolAmount(types2.FeePoolName, fee); err != nil {
		return nil, err
	}
	msg.SetSigner(address)
	return msg, nil
}

func (u *UtilityContext) HandleMessage(msg types2.Message) types.Error {
	switch x := msg.(type) {
	case *types2.MessageDoubleSign:
		return u.HandleMessageDoubleSign(x)
	case *types2.MessageSend:
		return u.HandleMessageSend(x)
	case *types2.MessageStakeFisherman:
		return u.HandleMessageStakeFisherman(x)
	case *types2.MessageEditStakeFisherman:
		return u.HandleMessageEditStakeFisherman(x)
	case *types2.MessageUnstakeFisherman:
		return u.HandleMessageUnstakeFisherman(x)
	case *types2.MessagePauseFisherman:
		return u.HandleMessagePauseFisherman(x)
	case *types2.MessageUnpauseFisherman:
		return u.HandleMessageUnpauseFisherman(x)
	case *types2.MessageFishermanPauseServiceNode:
		return u.HandleMessageFishermanPauseServiceNode(x)
	//case *types.MessageTestScore:
	//	return u.HandleMessageTestScore(x)
	//case *types.MessageProveTestScore:
	//	return u.HandleMessageProveTestScore(x)
	case *types2.MessageStakeApp:
		return u.HandleMessageStakeApp(x)
	case *types2.MessageEditStakeApp:
		return u.HandleMessageEditStakeApp(x)
	case *types2.MessageUnstakeApp:
		return u.HandleMessageUnstakeApp(x)
	case *types2.MessagePauseApp:
		return u.HandleMessagePauseApp(x)
	case *types2.MessageUnpauseApp:
		return u.HandleMessageUnpauseApp(x)
	case *types2.MessageStakeValidator:
		return u.HandleMessageStakeValidator(x)
	case *types2.MessageEditStakeValidator:
		return u.HandleMessageEditStakeValidator(x)
	case *types2.MessageUnstakeValidator:
		return u.HandleMessageUnstakeValidator(x)
	case *types2.MessagePauseValidator:
		return u.HandleMessagePauseValidator(x)
	case *types2.MessageUnpauseValidator:
		return u.HandleMessageUnpauseValidator(x)
	case *types2.MessageStakeServiceNode:
		return u.HandleMessageStakeServiceNode(x)
	case *types2.MessageEditStakeServiceNode:
		return u.HandleMessageEditStakeServiceNode(x)
	case *types2.MessageUnstakeServiceNode:
		return u.HandleMessageUnstakeServiceNode(x)
	case *types2.MessagePauseServiceNode:
		return u.HandleMessagePauseServiceNode(x)
	case *types2.MessageUnpauseServiceNode:
		return u.HandleMessageUnpauseServiceNode(x)
	case *types2.MessageChangeParameter:
		return u.HandleMessageChangeParameter(x)
	default:
		return types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetSignerCandidates(msg types2.Message) (candidates [][]byte, err types.Error) {
	switch x := msg.(type) {
	case *types2.MessageDoubleSign:
		return u.GetMessageDoubleSignSignerCandidates(x)
	case *types2.MessageSend:
		return u.GetMessageSendSignerCandidates(x)
	case *types2.MessageStakeFisherman:
		return u.GetMessageStakeFishermanSignerCandidates(x)
	case *types2.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanSignerCandidates(x)
	case *types2.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanSignerCandidates(x)
	case *types2.MessagePauseFisherman:
		return u.GetMessagePauseFishermanSignerCandidates(x)
	case *types2.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanSignerCandidates(x)
	case *types2.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeSignerCandidates(x)
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreSignerCandidates(x)
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreSignerCandidates(x)
	case *types2.MessageStakeApp:
		return u.GetMessageStakeAppSignerCandidates(x)
	case *types2.MessageEditStakeApp:
		return u.GetMessageEditStakeAppSignerCandidates(x)
	case *types2.MessageUnstakeApp:
		return u.GetMessageUnstakeAppSignerCandidates(x)
	case *types2.MessagePauseApp:
		return u.GetMessagePauseAppSignerCandidates(x)
	case *types2.MessageUnpauseApp:
		return u.GetMessageUnpauseAppSignerCandidates(x)
	case *types2.MessageStakeValidator:
		return u.GetMessageStakeValidatorSignerCandidates(x)
	case *types2.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorSignerCandidates(x)
	case *types2.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorSignerCandidates(x)
	case *types2.MessagePauseValidator:
		return u.GetMessagePauseValidatorSignerCandidates(x)
	case *types2.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorSignerCandidates(x)
	case *types2.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeSignerCandidates(x)
	case *types2.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeSignerCandidates(x)
	case *types2.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeSignerCandidates(x)
	case *types2.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeSignerCandidates(x)
	case *types2.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeSignerCandidates(x)
	case *types2.MessageChangeParameter:
		return u.GetMessageChangeParameterSignerCandidates(x)
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}
