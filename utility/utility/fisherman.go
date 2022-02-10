package utility

import (
	"pocket/utility/shared/crypto"
	"pocket/utility/utility/types"
	"pocket/shared/modules"
)

func (u *UtilityContext) HandleMessageTestScore(message *types.MessageTestScore) types.Error {
	// TODO
	panic("TODO")
}

func (u *UtilityContext) HandleMessageProveTestScore(message *types.MessageProveTestScore) types.Error {
	// TODO
	panic("TODO")
}

func (u *UtilityContext) HandleMessageStakeFisherman(message *types.MessageStakeFisherman) types.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	minStake, err := u.GetFishermanMinimumStake()
	if err != nil {
		return err
	}
	amount, err := StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	if BigIntLessThan(amount, minStake) {
		return types.ErrMinimumStake()
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmountError()
	}
	maxChains, err := u.GetFishermanMaxChains()
	if err != nil {
		return err
	}
	// validate chains
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(types.FishermanStakePoolName, amount); err != nil {
		return err
	}
	// ensure Fisherman doesn't already exist
	exists, err := u.GetFishermanExists(publicKey.Address())
	if err != nil {
		return err
	}
	if exists {
		return types.ErrAlreadyExists()
	}
	// insert the Fisherman structure
	if err := u.InsertFisherman(publicKey.Address(), message.PublicKey, message.OutputAddress, message.ServiceURL, message.Amount, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeFisherman(message *types.MessageEditStakeFisherman) types.Error {
	exists, err := u.GetFishermanExists(message.Address)
	if err != nil {
		return err
	}
	if !exists {
		return types.ErrNotExists()
	}
	amountToAdd, err := StringToBigInt(message.AmountToAdd)
	if err != nil {
		return err
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amountToAdd)
	if signerAccountAmount.Sign() == -1 {
		return types.ErrInsufficientAmountError()
	}
	maxChains, err := u.GetFishermanMaxChains()
	if err != nil {
		return err
	}
	// validate chains
	if len(message.Chains) > maxChains {
		return types.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(types.FishermanStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the Fisherman structure
	if err := u.UpdateFisherman(message.Address, message.ServiceURL, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeFisherman(message *types.MessageUnstakeFisherman) types.Error {
	status, err := u.GetFishermanStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != types.StakedStatus {
		return types.ErrInvalidStatus(status, types.StakedStatus)
	}
	unstakingHeight, err := u.CalculateFishermanUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetFishermanUnstakingHeightAndStatus(message.Address, unstakingHeight, types.UnstakingStatus); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeFishermansThatAreReady() types.Error {
	FishermansReadyToUnstake, err := u.GetFishermenReadyToUnstake()
	if err != nil {
		return err
	}
	for _, Fisherman := range FishermansReadyToUnstake {
		if err := u.SubPoolAmount(types.FishermanStakePoolName, Fisherman.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.AddAccountAmountString(Fisherman.GetOutputAddress(), Fisherman.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.DeleteFisherman(Fisherman.GetAddress()); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPausedFishermans() types.Error {
	maxPausedBlocks, err := u.GetFishermanMaxPausedBlocks()
	if err != nil {
		return err
	}
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	beforeHeight := latestHeight - int64(maxPausedBlocks)
	if beforeHeight < 0 { // genesis edge case
		beforeHeight = 0
	}
	if err := u.UnstakeFishermansPausedBefore(beforeHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessagePauseFisherman(message *types.MessagePauseFisherman) types.Error {
	height, err := u.GetFishermanPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if height != 0 {
		return types.ErrAlreadyPaused()
	}
	height, err = u.GetLatestHeight()
	if err != nil {
		return err
	}
	if err := u.SetFishermanPauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageFishermanPauseServiceNode(message *types.MessageFishermanPauseServiceNode) types.Error {
	exists, err := u.GetFishermanExists(message.Reporter)
	if err != nil {
		return err
	}
	if !exists {
		return types.ErrNotExists()
	}
	height, err := u.GetServiceNodePauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if height != 0 {
		return types.ErrAlreadyPaused()
	}
	height, err = u.GetLatestHeight()
	if err != nil {
		return err
	}
	if err := u.SetServiceNodePauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnpauseFisherman(message *types.MessageUnpauseFisherman) types.Error {
	pausedHeight, err := u.GetFishermanPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == 0 {
		return types.ErrNotPaused()
	}
	minPauseBlocks, err := u.GetFishermanMinimumPauseBlocks()
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
	if err := u.SetFishermanPauseHeight(message.Address, types.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetFishermanExists(address []byte) (exists bool, err types.Error) {
	store := u.Store()
	exists, er := store.GetFishermanExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertFisherman(address, publicKey, output []byte, serviceURL, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.InsertFisherman(address, publicKey, output, false, types.StakedStatus, serviceURL, amount, chains, types.ZeroInt, types.ZeroInt)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) UpdateFisherman(address []byte, serviceURL, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.UpdateFisherman(address, serviceURL, amount, chains)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteFisherman(address []byte) types.Error {
	store := u.Store()
	if err := store.DeleteFisherman(address); err != nil {
		return types.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetFishermenReadyToUnstake() (Fishermans []modules.UnstakingActor, err types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingFishermans, er := store.GetFishermanReadyToUnstake(latestHeight, types.UnstakingStatus)
	if er != nil {
		return nil, types.ErrGetReadyToUnstake(er)
	}
	return unstakingFishermans, nil
}

func (u *UtilityContext) UnstakeFishermansPausedBefore(pausedBeforeHeight int64) types.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateFishermanUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetFishermansStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, types.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetFishermanStatus(address []byte) (status int, err types.Error) {
	store := u.Store()
	status, er := store.GetFishermanStatus(address)
	if er != nil {
		return types.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) (err types.Error) {
	store := u.Store()
	if er := store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, status); er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetFishermanPauseHeightIfExists(address []byte) (FishermanPauseHeight int64, err types.Error) {
	store := u.Store()
	FishermanPauseHeight, er := store.GetFishermanPauseHeightIfExists(address)
	if er != nil {
		return types.ZeroInt, types.ErrGetPauseHeight(er)
	}
	return FishermanPauseHeight, nil
}

func (u *UtilityContext) SetFishermanPauseHeight(address []byte, height int64) types.Error {
	store := u.Store()
	if err := store.SetFishermanPauseHeight(address, height); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateFishermanUnstakingHeight() (unstakingHeight int64, err types.Error) {
	unstakingBlocks, err := u.GetFishermanUnstakingBlocks()
	if err != nil {
		return types.ZeroInt, err
	}
	unstakingHeight, err = u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return types.ZeroInt, err
	}
	return
}

func (u *UtilityContext) GetMessageStakeFishermanSignerCandidates(msg *types.MessageStakeFisherman) (candidates [][]byte, err types.Error) {
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, pk.Address())
	return
}

func (u *UtilityContext) GetMessageEditStakeFishermanSignerCandidates(msg *types.MessageEditStakeFisherman) (candidates [][]byte, err types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnstakeFishermanSignerCandidates(msg *types.MessageUnstakeFisherman) (candidates [][]byte, err types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnpauseFishermanSignerCandidates(msg *types.MessageUnpauseFisherman) (candidates [][]byte, err types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessagePauseFishermanSignerCandidates(msg *types.MessagePauseFisherman) (candidates [][]byte, err types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeSignerCandidates(msg *types.MessageFishermanPauseServiceNode) (candidates [][]byte, err types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Reporter)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Reporter)
	return
}

func (u *UtilityContext) GetFishermanOutputAddress(operator []byte) (output []byte, err types.Error) {
	store := u.Store()
	output, er := store.GetFishermanOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
