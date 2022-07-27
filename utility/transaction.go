package utility

import (
	"bytes"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
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
	fee, err := u.GetFee(msg, msg.GetActorType())
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
		return nil, types.ErrInsufficientAmount()
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
	if err := u.AddPoolAmount(typesGenesis.FeePoolName, fee); err != nil {
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
	case *typesUtil.MessageStake:
		return u.HandleStakeMessage(x)
	case *typesUtil.MessageEditStake:
		return u.HandleEditStakeMessage(x)
	case *typesUtil.MessageUnstake:
		return u.HandleUnstakeMessage(x)
	case *typesUtil.MessageUnpause:
		return u.HandleUnpauseMessage(x)
	case *typesUtil.MessageChangeParameter:
		return u.HandleMessageChangeParameter(x)
	default:
		return types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) HandleMessageSend(message *typesUtil.MessageSend) types.Error {
	// convert the amount to big.Int
	amount, err := types.StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	// get the sender's account amount
	fromAccountAmount, err := u.GetAccountAmount(message.FromAddress)
	if err != nil {
		return err
	}
	// subtract that amount from the sender
	fromAccountAmount.Sub(fromAccountAmount, amount)
	// if they go negative, they don't have sufficient funds
	// NOTE: we don't use the u.SubtractAccountAmount() function because Utility needs to do this check
	if fromAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmount()
	}
	// add the amount to the recipient's account
	if err = u.AddAccountAmount(message.ToAddress, amount); err != nil {
		return err
	}
	// set the sender's account amount
	if err = u.SetAccountAmount(message.FromAddress, fromAccountAmount); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleStakeMessage(message *typesUtil.MessageStake) types.Error {
	publicKey, err := u.BytesToPublicKey(message.PublicKey)
	if err != nil {
		return err
	}
	// ensure above minimum stake
	amount, err := u.CheckAboveMinStake(message.ActorType, message.Amount)
	if err != nil {
		return err
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	// calculate new signer account amount
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmount()
	}
	// validators don't have chains field
	if err = u.CheckBelowMaxChains(message.ActorType, message.Chains); err != nil {
		return err
	}
	// ensure actor doesn't already exist
	if exists, err := u.GetActorExists(message.ActorType, publicKey.Address()); err != nil || exists {
		if exists {
			return types.ErrAlreadyExists()
		}
		return err
	}
	// update account amount
	if err = u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err = u.AddPoolAmount(typesGenesis.AppStakePoolName, amount); err != nil {
		return err
	}
	var er error
	store := u.Store()
	// insert actor
	switch message.ActorType {
	case typesUtil.ActorType_App:
		maxRelays, err := u.CalculateAppRelays(message.Amount)
		if err != nil {
			return err
		}
		er = store.InsertApp(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, typesUtil.StakedStatus, maxRelays, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case typesUtil.ActorType_Fish:
		er = store.InsertFisherman(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, typesUtil.StakedStatus, *message.ServiceUrl, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case typesUtil.ActorType_Node:
		er = store.InsertServiceNode(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, typesUtil.StakedStatus, *message.ServiceUrl, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case typesUtil.ActorType_Val:
		er = store.InsertValidator(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, typesUtil.StakedStatus, *message.ServiceUrl, message.Amount, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	}
	if er != nil {
		return types.ErrInsert(er)
	}
	return nil
}

func (u *UtilityContext) HandleEditStakeMessage(message *typesUtil.MessageEditStake) types.Error {
	// ensure actor exists
	if exists, err := u.GetActorExists(message.ActorType, message.Address); err != nil || !exists {
		if !exists {
			return types.ErrNotExists()
		}
		return err
	}
	currentStakeAmount, err := u.GetStakeAmount(message.ActorType, message.Address)
	if err != nil {
		return err
	}
	amount, err := types.StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	// ensure new stake >= current stake
	amount.Sub(amount, currentStakeAmount)
	if amount.Sign() == -1 {
		return types.ErrStakeLess()
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmount()
	}
	if err = u.CheckBelowMaxChains(message.ActorType, message.Chains); err != nil {
		return err
	}
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(typesGenesis.AppStakePoolName, amount); err != nil {
		return err
	}
	store := u.Store()
	var er error
	switch message.ActorType {
	case typesUtil.ActorType_App:
		maxRelays, err := u.CalculateAppRelays(message.Amount)
		if err != nil {
			return err
		}
		er = store.UpdateApp(message.Address, maxRelays, message.Amount, message.Chains)
	case typesUtil.ActorType_Fish:
		er = store.UpdateFisherman(message.Address, *message.ServiceUrl, message.Amount, message.Chains)
	case typesUtil.ActorType_Node:
		er = store.UpdateServiceNode(message.Address, *message.ServiceUrl, message.Amount, message.Chains)
	case typesUtil.ActorType_Val:
		er = store.UpdateValidator(message.Address, *message.ServiceUrl, message.Amount)
	}
	if er != nil {
		return types.ErrInsert(er)
	}
	return nil
}

func (u *UtilityContext) HandleUnstakeMessage(message *typesUtil.MessageUnstake) types.Error {
	if status, err := u.GetActorStatus(message.ActorType, message.Address); err != nil || status != typesUtil.StakedStatus {
		if status != typesUtil.StakedStatus {
			return types.ErrInvalidStatus(status, typesUtil.StakedStatus)
		}
		return err
	}
	unstakingHeight, err := u.GetUnstakingHeight(message.ActorType)
	if err != nil {
		return err
	}
	if err = u.SetActorUnstaking(message.ActorType, unstakingHeight, message.Address); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleUnpauseMessage(message *typesUtil.MessageUnpause) types.Error {
	pausedHeight, err := u.GetPauseHeight(message.ActorType, message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == typesUtil.HeightNotUsed {
		return types.ErrNotPaused()
	}
	minPauseBlocks, err := u.GetMinimumPauseBlocks(message.ActorType)
	if err != nil {
		return err
	}
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	if latestHeight < int64(minPauseBlocks)+pausedHeight {
		return types.ErrNotReadyToUnpause()
	}
	if err = u.SetActorPauseHeight(message.ActorType, message.Address, types.HeightNotUsed); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageDoubleSign(message *typesUtil.MessageDoubleSign) types.Error {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	evidenceAge := latestHeight - message.VoteA.Height
	maxEvidenceAge, err := u.GetMaxEvidenceAgeInBlocks()
	if err != nil {
		return err
	}
	if evidenceAge > int64(maxEvidenceAge) {
		return types.ErrMaxEvidenceAge()
	}
	pk, er := crypto.NewPublicKeyFromBytes(message.VoteB.PublicKey)
	if er != nil {
		return types.ErrNewPublicKeyFromBytes(er)
	}
	doubleSigner := pk.Address()
	// burn validator for double signing blocks
	burnPercentage, err := u.GetDoubleSignBurnPercentage()
	if err != nil {
		return err
	}
	if err := u.BurnActor(typesUtil.ActorType_Val, burnPercentage, doubleSigner); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageChangeParameter(message *typesUtil.MessageChangeParameter) types.Error {
	cdc := u.Codec()
	v, err := cdc.FromAny(message.ParameterValue)
	if err != nil {
		return types.ErrProtoFromAny(err)
	}
	return u.UpdateParam(message.ParameterKey, v)
}

func (u *UtilityContext) GetSignerCandidates(msg typesUtil.Message) ([][]byte, types.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignSignerCandidates(x)
	case *typesUtil.MessageSend:
		return u.GetMessageSendSignerCandidates(x)
	case *typesUtil.MessageStake:
		return u.GetMessageStakeSignerCandidates(x)
	case *typesUtil.MessageUnstake:
		return u.GetMessageUnstakeSignerCandidates(x)
	case *typesUtil.MessageUnpause:
		return u.GetMessageUnpauseSignerCandidates(x)
	case *typesUtil.MessageChangeParameter:
		return u.GetMessageChangeParameterSignerCandidates(x)
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetMessageStakeSignerCandidates(msg *typesUtil.MessageStake) ([][]byte, types.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress)
	candidates = append(candidates, pk.Address())
	return candidates, nil
}

func (u *UtilityContext) GetMessageEditStakeSignerCandidates(msg *typesUtil.MessageEditStake) ([][]byte, types.Error) {
	output, err := u.GetActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnstakeSignerCandidates(msg *typesUtil.MessageUnstake) ([][]byte, types.Error) {
	output, err := u.GetActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnpauseSignerCandidates(msg *typesUtil.MessageUnpause) ([][]byte, types.Error) {
	output, err := u.GetActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageSendSignerCandidates(msg *typesUtil.MessageSend) ([][]byte, types.Error) {
	return [][]byte{msg.FromAddress}, nil
}

func (u *UtilityContext) GetMessageDoubleSignSignerCandidates(msg *typesUtil.MessageDoubleSign) ([][]byte, types.Error) {
	return [][]byte{msg.ReporterAddress}, nil
}
