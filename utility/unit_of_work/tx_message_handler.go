package unit_of_work

import (
	"encoding/hex"
	"math/big"

	ibcTypes "github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// handleMessage handles the message by applying the underlying business logic associated with it.
func (u *baseUtilityUnitOfWork) handleMessage(msg typesUtil.Message) (err coreTypes.Error) {
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
	case *ibcTypes.UpdateIBCStore:
		return u.handleUpdateIBCStore(x)
	case *ibcTypes.PruneIBCStore:
		return u.handlePruneIBCStore(x)
	case *typesUtil.MessageUpgrade:
		return u.handleMessageUpgrade(x)
	// TODO: 0xbigboss MessageCancelUpgrade
	default:
		return coreTypes.ErrUnknownMessage(x)
	}
}

func (u *baseUtilityUnitOfWork) handleMessageSend(message *typesUtil.MessageSend) coreTypes.Error {
	// convert the amount to big.Int
	amount, er := utils.StringToBigInt(message.Amount)
	if er != nil {
		return coreTypes.ErrStringToBigInt(er)
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
		return coreTypes.ErrInsufficientAmount(hex.EncodeToString(message.FromAddress))
	}
	// add the amount to the recipient's account
	if err := u.addAccountAmount(message.ToAddress, amount); err != nil {
		return err
	}
	// set the sender's account amount
	if err := u.setAccountAmount(message.FromAddress, fromAccountAmount); err != nil {
		return err
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handleStakeMessage(message *typesUtil.MessageStake) coreTypes.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return coreTypes.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	amount, err := u.checkAboveMinStake(message.ActorType, message.Amount)
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
		return coreTypes.ErrInsufficientAmount(hex.EncodeToString(message.Signer))
	}
	// validators don't have chains field
	if err := u.checkBelowMaxChains(message.ActorType, message.Chains); err != nil {
		return err
	}
	// ensure actor doesn't already exist
	if exists, err := u.getActorExists(message.ActorType, publicKey.Address()); err != nil || exists {
		if exists {
			return coreTypes.ErrAlreadyExists()
		}
		return err
	}
	// update account amount
	if err := u.setAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_APP_STAKE.Address(), amount); err != nil {
		return err
	}

	// insert actor
	switch message.ActorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = u.persistenceRWContext.InsertApp(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(coreTypes.StakeStatus_Staked), message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = u.persistenceRWContext.InsertFisherman(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(coreTypes.StakeStatus_Staked), message.ServiceUrl, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		er = u.persistenceRWContext.InsertServicer(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(coreTypes.StakeStatus_Staked), message.ServiceUrl, message.Amount, message.Chains, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = u.persistenceRWContext.InsertValidator(publicKey.Address(), publicKey.Bytes(), message.OutputAddress, false, int32(coreTypes.StakeStatus_Staked), message.ServiceUrl, message.Amount, typesUtil.HeightNotUsed, typesUtil.HeightNotUsed)
	}
	if er != nil {
		return coreTypes.ErrInsert(er)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handleEditStakeMessage(message *typesUtil.MessageEditStake) coreTypes.Error {
	// ensure actor exists
	if exists, err := u.getActorExists(message.ActorType, message.Address); err != nil || !exists {
		if !exists {
			return coreTypes.ErrNotExists()
		}
		return err
	}
	currentStakeAmount, err := u.getActorStakeAmount(message.ActorType, message.Address)
	if err != nil {
		return err
	}
	amount, er := utils.StringToBigInt(message.Amount)
	if er != nil {
		return coreTypes.ErrStringToBigInt(err)
	}
	// ensure new stake >= current stake
	amount.Sub(amount, currentStakeAmount)
	if amount.Sign() == -1 {
		return coreTypes.ErrStakeLess()
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.getAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return coreTypes.ErrInsufficientAmount(hex.EncodeToString(message.Signer))
	}
	if err := u.checkBelowMaxChains(message.ActorType, message.Chains); err != nil {
		return err
	}
	// update account amount
	if err := u.setAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_APP_STAKE.Address(), amount); err != nil {
		return err
	}
	switch message.ActorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = u.persistenceRWContext.UpdateApp(message.Address, message.Amount, message.Chains)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = u.persistenceRWContext.UpdateFisherman(message.Address, message.ServiceUrl, message.Amount, message.Chains)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		er = u.persistenceRWContext.UpdateServicer(message.Address, message.ServiceUrl, message.Amount, message.Chains)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = u.persistenceRWContext.UpdateValidator(message.Address, message.ServiceUrl, message.Amount)
	}
	if er != nil {
		return coreTypes.ErrInsert(er)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handleUnstakeMessage(message *typesUtil.MessageUnstake) coreTypes.Error {
	if status, err := u.getActorStatus(message.ActorType, message.Address); err != nil || status != coreTypes.StakeStatus_Staked {
		if status != coreTypes.StakeStatus_Staked {
			return coreTypes.ErrInvalidStatus(status, coreTypes.StakeStatus_Staked)
		}
		return err
	}
	unbondingHeight, err := u.getUnbondingHeight(message.ActorType)
	if err != nil {
		return err
	}
	if err := u.setActorUnbondingHeight(message.ActorType, message.Address, unbondingHeight); err != nil {
		return err
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handleUnpauseMessage(message *typesUtil.MessageUnpause) coreTypes.Error {
	pausedHeight, err := u.getPausedHeightIfExists(message.ActorType, message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == typesUtil.HeightNotUsed {
		return coreTypes.ErrNotPaused()
	}
	minPauseBlocks, err := u.getMinRequiredPausedBlocks(message.ActorType)
	if err != nil {
		return err
	}
	if u.height < int64(minPauseBlocks)+pausedHeight {
		return coreTypes.ErrNotReadyToUnpause()
	}
	if err := u.setActorPausedHeight(message.ActorType, message.Address, typesUtil.HeightNotUsed); err != nil {
		return err
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handleMessageChangeParameter(message *typesUtil.MessageChangeParameter) coreTypes.Error {
	v, err := codec.GetCodec().FromAny(message.ParameterValue)
	if err != nil {
		return coreTypes.ErrProtoFromAny(err)
	}
	return u.updateParam(message.ParameterKey, v)
}

func (u *baseUtilityUnitOfWork) handleUpdateIBCStore(message *ibcTypes.UpdateIBCStore) coreTypes.Error {
	if err := u.persistenceRWContext.SetIBCStoreEntry(message.Key, message.Value); err != nil {
		return coreTypes.ErrIBCUpdatingStore(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handlePruneIBCStore(message *ibcTypes.PruneIBCStore) coreTypes.Error {
	if err := u.persistenceRWContext.SetIBCStoreEntry(message.Key, nil); err != nil {
		return coreTypes.ErrIBCUpdatingStore(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) handleMessageUpgrade(message *typesUtil.MessageUpgrade) coreTypes.Error {
	u.logger.Info().Str("version", message.Version).Int64("height", message.Height).Msg("setting upgrade")
	if err := u.persistenceRWContext.SetUpgrade(message.Version, message.Height); err != nil {
		return coreTypes.ErrSettingUpgrade(err)
	}
	return nil
}

// TODO: 0xbigboss MessageCancelUpgrade

func (u *baseUtilityUnitOfWork) checkBelowMaxChains(actorType coreTypes.ActorType, chains []string) coreTypes.Error {
	// validators don't have chains field
	if actorType == coreTypes.ActorType_ACTOR_TYPE_VAL {
		return nil
	}

	maxChains, err := u.getMaxAllowedChains(actorType)
	if err != nil {
		return err
	}
	if len(chains) > maxChains {
		return coreTypes.ErrMaxChains(maxChains)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) checkAboveMinStake(actorType coreTypes.ActorType, amountStr string) (*big.Int, coreTypes.Error) {
	minStake, err := u.getMinRequiredStakeAmount(actorType)
	if err != nil {
		return nil, err
	}
	amount, er := utils.StringToBigInt(amountStr)
	if er != nil {
		return nil, coreTypes.ErrStringToBigInt(err)
	}
	if utils.BigIntLessThan(amount, minStake) {
		return nil, coreTypes.ErrMinimumStake()
	}
	return amount, nil
}
