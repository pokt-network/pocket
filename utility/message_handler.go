package utility

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	TransactionGossipMessageContentType = "utility.TransactionGossipMessage"
)

func (u *utilityModule) HandleMessage(message *anypb.Any) error {
	switch message.MessageName() {
	case TransactionGossipMessageContentType:
		msg, err := codec.GetCodec().FromAny(message)
		if err != nil {
			return err
		}

		if transactionGossipMsg, ok := msg.(*types.TransactionGossipMessage); !ok {
			return fmt.Errorf("failed to cast message to UtilityMessage")
		} else if err := u.CheckTransaction(transactionGossipMsg.Tx); err != nil {
			return err
		}

		log.Println("MEMPOOL: Successfully added a new message to the mempool!")

	default:
		return types.ErrUnknownMessageType(message.MessageName())
	}

	return nil
}

func (u *utilityContext) handleMessage(msg typesUtil.Message) (err typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageSend:
		return u.handleMessageSend(x)
	case *typesUtil.MessageStake:
		return u.handleStakeMessage(x)
	case *typesUtil.MessageEditStake:
		return u.handleEditStakeMessage(x)
	case *typesUtil.MessageUnstake:
		return u.handleUnstakeMessage(x)
	case *typesUtil.MessageUnpause:
		return u.handleUnpauseMessage(x)
	case *typesUtil.MessageChangeParameter:
		return u.handleMessageChangeParameter(x)
	default:
		return typesUtil.ErrUnknownMessage(x)
	}
}

func (u *utilityContext) handleMessageSend(message *typesUtil.MessageSend) typesUtil.Error {
	// convert the amount to big.Int
	amount, er := converters.StringToBigInt(message.Amount)
	if er != nil {
		return typesUtil.ErrStringToBigInt(er)
	}
	// get the sender's account amount
	fromAccountAmount, err := u.getAccountAmount(message.FromAddress)
	if err != nil {
		return err
	}
	// subtract that amount from the sender
	fromAccountAmount.Sub(fromAccountAmount, amount)
	// if they go negative, they don't have sufficient funds
	// NOTE: we don't use the u.SubtractAccountAmount() function because Utility needs to do this check
	if fromAccountAmount.Sign() == -1 {
		return typesUtil.ErrInsufficientAmount(hex.EncodeToString(message.FromAddress))
	}
	// add the amount to the recipient's account
	if err = u.addAccountAmount(message.ToAddress, amount); err != nil {
		return err
	}
	// set the sender's account amount
	if err = u.setAccountAmount(message.FromAddress, fromAccountAmount); err != nil {
		return err
	}
	return nil
}

func (u *utilityContext) handleStakeMessage(message *typesUtil.MessageStake) typesUtil.Error {
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
	signerAccountAmount, err := u.getAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	// calculate new signer account amount
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return typesUtil.ErrInsufficientAmount(hex.EncodeToString(message.Signer))
	}
	// validators don't have chains field
	if err = u.CheckBelowMaxChains(message.ActorType, message.Chains); err != nil {
		return err
	}
	// ensure actor doesn't already exist
	if exists, err := u.GetActorExists(message.ActorType, publicKey.Address()); err != nil || exists {
		if exists {
			return typesUtil.ErrAlreadyExists()
		}
		return err
	}
	// update account amount
	if err = u.setAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err = u.addPoolAmount(coreTypes.Pools_POOLS_APP_STAKE.FriendlyName(), amount); err != nil {
		return err
	}
	var er error
	store := u.Store()
	// insert actor
	switch message.ActorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		maxRelays, err := u.CalculateAppRelays(message.Amount)
		if err != nil {
			return err
		}
		er = store.InsertApp(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(typesUtil.StakeStatus_Staked), maxRelays, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = store.InsertFisherman(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(typesUtil.StakeStatus_Staked), message.ServiceUrl, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		er = store.InsertServiceNode(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(typesUtil.StakeStatus_Staked), message.ServiceUrl, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = store.InsertValidator(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(typesUtil.StakeStatus_Staked), message.ServiceUrl, message.Amount, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	}
	if er != nil {
		return typesUtil.ErrInsert(er)
	}
	return nil
}

func (u *utilityContext) handleEditStakeMessage(message *typesUtil.MessageEditStake) typesUtil.Error {
	// ensure actor exists
	if exists, err := u.GetActorExists(message.ActorType, message.Address); err != nil || !exists {
		if !exists {
			return typesUtil.ErrNotExists()
		}
		return err
	}
	currentStakeAmount, err := u.GetStakeAmount(message.ActorType, message.Address)
	if err != nil {
		return err
	}
	amount, er := converters.StringToBigInt(message.Amount)
	if er != nil {
		return typesUtil.ErrStringToBigInt(err)
	}
	// ensure new stake >= current stake
	amount.Sub(amount, currentStakeAmount)
	if amount.Sign() == -1 {
		return typesUtil.ErrStakeLess()
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.getAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return typesUtil.ErrInsufficientAmount(hex.EncodeToString(message.Signer))
	}
	if err = u.CheckBelowMaxChains(message.ActorType, message.Chains); err != nil {
		return err
	}
	// update account amount
	if err := u.setAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_APP_STAKE.FriendlyName(), amount); err != nil {
		return err
	}
	store := u.Store()
	switch message.ActorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		maxRelays, err := u.CalculateAppRelays(message.Amount)
		if err != nil {
			return err
		}
		er = store.UpdateApp(message.Address, maxRelays, message.Amount, message.Chains)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = store.UpdateFisherman(message.Address, message.ServiceUrl, message.Amount, message.Chains)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		er = store.UpdateServiceNode(message.Address, message.ServiceUrl, message.Amount, message.Chains)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = store.UpdateValidator(message.Address, message.ServiceUrl, message.Amount)
	}
	if er != nil {
		return typesUtil.ErrInsert(er)
	}
	return nil
}

func (u *utilityContext) handleUnstakeMessage(message *typesUtil.MessageUnstake) typesUtil.Error {
	if status, err := u.getActorStatus(message.ActorType, message.Address); err != nil || status != int32(typesUtil.StakeStatus_Staked) {
		if status != int32(typesUtil.StakeStatus_Staked) {
			return typesUtil.ErrInvalidStatus(status, int32(typesUtil.StakeStatus_Staked))
		}
		return err
	}
	unstakingHeight, err := u.GetUnstakingHeight(message.ActorType)
	if err != nil {
		return err
	}
	if err = u.setActorUnstakingHeight(message.ActorType, message.Address, unstakingHeight); err != nil {
		return err
	}
	return nil
}

func (u *utilityContext) handleUnpauseMessage(message *typesUtil.MessageUnpause) typesUtil.Error {
	pausedHeight, err := u.getPausedHeightIfExists(message.ActorType, message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == typesUtil.HeightNotUsed {
		return typesUtil.ErrNotPaused()
	}
	minPauseBlocks, err := u.getMinRequiredPausedBlocks(message.ActorType)
	if err != nil {
		return err
	}
	latestHeight, err := u.GetLatestBlockHeight()
	if err != nil {
		return err
	}
	if latestHeight < int64(minPauseBlocks)+pausedHeight {
		return typesUtil.ErrNotReadyToUnpause()
	}
	if err = u.setActorPausedHeight(message.ActorType, message.Address, typesUtil.HeightNotUsed); err != nil {
		return err
	}
	return nil
}

func (u *utilityContext) handleMessageChangeParameter(message *typesUtil.MessageChangeParameter) typesUtil.Error {
	v, err := codec.GetCodec().FromAny(message.ParameterValue)
	if err != nil {
		return typesUtil.ErrProtoFromAny(err)
	}
	return u.updateParam(message.ParameterKey, v)
}

func (u *utilityContext) GetSignerCandidates(msg typesUtil.Message) ([][]byte, typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageSend:
		return u.GetMessageSendSignerCandidates(x)
	case *typesUtil.MessageStake:
		return u.GetMessageStakeSignerCandidates(x)
	case *typesUtil.MessageUnstake:
		return u.GetMessageUnstakeSignerCandidates(x)
	case *typesUtil.MessageUnpause:
		return u.getMessageUnpauseSignerCandidates(x)
	case *typesUtil.MessageChangeParameter:
		return u.GetMessageChangeParameterSignerCandidates(x)
	default:
		return nil, typesUtil.ErrUnknownMessage(x)
	}
}

func (u *utilityContext) GetMessageStakeSignerCandidates(msg *typesUtil.MessageStake) ([][]byte, typesUtil.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, typesUtil.ErrNewPublicKeyFromBytes(er)
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress)
	candidates = append(candidates, pk.Address())
	return candidates, nil
}

func (u *utilityContext) GetMessageEditStakeSignerCandidates(msg *typesUtil.MessageEditStake) ([][]byte, typesUtil.Error) {
	output, err := u.GetActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *utilityContext) GetMessageUnstakeSignerCandidates(msg *typesUtil.MessageUnstake) ([][]byte, typesUtil.Error) {
	output, err := u.GetActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *utilityContext) getMessageUnpauseSignerCandidates(msg *typesUtil.MessageUnpause) ([][]byte, typesUtil.Error) {
	output, err := u.GetActorOutputAddress(msg.ActorType, msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *utilityContext) GetMessageSendSignerCandidates(msg *typesUtil.MessageSend) ([][]byte, typesUtil.Error) {
	return [][]byte{msg.FromAddress}, nil
}
