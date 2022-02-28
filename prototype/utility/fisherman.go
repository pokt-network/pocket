package utility

import (
	"pocket/shared/crypto"
	"pocket/shared/modules"
	types2 "pocket/utility/types"
)

func (u *UtilityContext) HandleMessageTestScore(message *types2.MessageTestScore) types2.Error {
	// TODO
	panic("TODO")
}

func (u *UtilityContext) HandleMessageProveTestScore(message *types2.MessageProveTestScore) types2.Error {
	// TODO
	panic("TODO")
}

func (u *UtilityContext) HandleMessageStakeFisherman(message *types2.MessageStakeFisherman) types2.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types2.ErrNewPublicKeyFromBytes(er)
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
		return types2.ErrMinimumStake()
	}
	// ensure signer has sufficient funding for the stake
	signerAccountAmount, err := u.GetAccountAmount(message.Signer)
	if err != nil {
		return err
	}
	signerAccountAmount.Sub(signerAccountAmount, amount)
	if signerAccountAmount.Sign() == -1 {
		return types2.ErrInsufficientAmountError()
	}
	maxChains, err := u.GetFishermanMaxChains()
	if err != nil {
		return err
	}
	// validate chains
	if len(message.Chains) > maxChains {
		return types2.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(types2.FishermanStakePoolName, amount); err != nil {
		return err
	}
	// ensure Fisherman doesn't already exist
	exists, err := u.GetFishermanExists(publicKey.Address())
	if err != nil {
		return err
	}
	if exists {
		return types2.ErrAlreadyExists()
	}
	// insert the Fisherman structure
	if err := u.InsertFisherman(publicKey.Address(), message.PublicKey, message.OutputAddress, message.ServiceURL, message.Amount, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeFisherman(message *types2.MessageEditStakeFisherman) types2.Error {
	exists, err := u.GetFishermanExists(message.Address)
	if err != nil {
		return err
	}
	if !exists {
		return types2.ErrNotExists()
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
		return types2.ErrInsufficientAmountError()
	}
	maxChains, err := u.GetFishermanMaxChains()
	if err != nil {
		return err
	}
	// validate chains
	if len(message.Chains) > maxChains {
		return types2.ErrMaxChains(maxChains)
	}
	// update account amount
	if err := u.SetAccount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(types2.FishermanStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the Fisherman structure
	if err := u.UpdateFisherman(message.Address, message.ServiceURL, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeFisherman(message *types2.MessageUnstakeFisherman) types2.Error {
	status, err := u.GetFishermanStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != types2.StakedStatus {
		return types2.ErrInvalidStatus(status, types2.StakedStatus)
	}
	unstakingHeight, err := u.CalculateFishermanUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetFishermanUnstakingHeightAndStatus(message.Address, unstakingHeight, types2.UnstakingStatus); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeFishermansThatAreReady() types2.Error {
	FishermansReadyToUnstake, err := u.GetFishermenReadyToUnstake()
	if err != nil {
		return err
	}
	for _, Fisherman := range FishermansReadyToUnstake {
		if err := u.SubPoolAmount(types2.FishermanStakePoolName, Fisherman.GetStakeAmount()); err != nil {
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

func (u *UtilityContext) BeginUnstakingMaxPausedFishermans() types2.Error {
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

func (u *UtilityContext) HandleMessagePauseFisherman(message *types2.MessagePauseFisherman) types2.Error {
	height, err := u.GetFishermanPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if height != 0 {
		return types2.ErrAlreadyPaused()
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

func (u *UtilityContext) HandleMessageFishermanPauseServiceNode(message *types2.MessageFishermanPauseServiceNode) types2.Error {
	exists, err := u.GetFishermanExists(message.Reporter)
	if err != nil {
		return err
	}
	if !exists {
		return types2.ErrNotExists()
	}
	height, err := u.GetServiceNodePauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if height != 0 {
		return types2.ErrAlreadyPaused()
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

func (u *UtilityContext) HandleMessageUnpauseFisherman(message *types2.MessageUnpauseFisherman) types2.Error {
	pausedHeight, err := u.GetFishermanPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == 0 {
		return types2.ErrNotPaused()
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
		return types2.ErrNotReadyToUnpause()
	}
	if err := u.SetFishermanPauseHeight(message.Address, types2.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetFishermanExists(address []byte) (exists bool, err types2.Error) {
	store := u.Store()
	exists, er := store.GetFishermanExists(address)
	if er != nil {
		return false, types2.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertFisherman(address, publicKey, output []byte, serviceURL, amount string, chains []string) types2.Error {
	store := u.Store()
	err := store.InsertFisherman(address, publicKey, output, false, types2.StakedStatus, serviceURL, amount, chains, types2.ZeroInt, types2.ZeroInt)
	if err != nil {
		return types2.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) UpdateFisherman(address []byte, serviceURL, amount string, chains []string) types2.Error {
	store := u.Store()
	err := store.UpdateFisherman(address, serviceURL, amount, chains)
	if err != nil {
		return types2.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteFisherman(address []byte) types2.Error {
	store := u.Store()
	if err := store.DeleteFisherman(address); err != nil {
		return types2.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetFishermenReadyToUnstake() (Fishermans []modules.UnstakingActor, err types2.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingFishermans, er := store.GetFishermanReadyToUnstake(latestHeight, types2.UnstakingStatus)
	if er != nil {
		return nil, types2.ErrGetReadyToUnstake(er)
	}
	return unstakingFishermans, nil
}

func (u *UtilityContext) UnstakeFishermansPausedBefore(pausedBeforeHeight int64) types2.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateFishermanUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetFishermansStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, types2.UnstakingStatus)
	if er != nil {
		return types2.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetFishermanStatus(address []byte) (status int, err types2.Error) {
	store := u.Store()
	status, er := store.GetFishermanStatus(address)
	if er != nil {
		return types2.ZeroInt, types2.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) (err types2.Error) {
	store := u.Store()
	if er := store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, status); er != nil {
		return types2.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetFishermanPauseHeightIfExists(address []byte) (FishermanPauseHeight int64, err types2.Error) {
	store := u.Store()
	FishermanPauseHeight, er := store.GetFishermanPauseHeightIfExists(address)
	if er != nil {
		return types2.ZeroInt, types2.ErrGetPauseHeight(er)
	}
	return FishermanPauseHeight, nil
}

func (u *UtilityContext) SetFishermanPauseHeight(address []byte, height int64) types2.Error {
	store := u.Store()
	if err := store.SetFishermanPauseHeight(address, height); err != nil {
		return types2.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateFishermanUnstakingHeight() (unstakingHeight int64, err types2.Error) {
	unstakingBlocks, err := u.GetFishermanUnstakingBlocks()
	if err != nil {
		return types2.ZeroInt, err
	}
	unstakingHeight, err = u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return types2.ZeroInt, err
	}
	return
}

func (u *UtilityContext) GetMessageStakeFishermanSignerCandidates(msg *types2.MessageStakeFisherman) (candidates [][]byte, err types2.Error) {
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types2.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, pk.Address())
	return
}

func (u *UtilityContext) GetMessageEditStakeFishermanSignerCandidates(msg *types2.MessageEditStakeFisherman) (candidates [][]byte, err types2.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnstakeFishermanSignerCandidates(msg *types2.MessageUnstakeFisherman) (candidates [][]byte, err types2.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnpauseFishermanSignerCandidates(msg *types2.MessageUnpauseFisherman) (candidates [][]byte, err types2.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessagePauseFishermanSignerCandidates(msg *types2.MessagePauseFisherman) (candidates [][]byte, err types2.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeSignerCandidates(msg *types2.MessageFishermanPauseServiceNode) (candidates [][]byte, err types2.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Reporter)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Reporter)
	return
}

func (u *UtilityContext) GetFishermanOutputAddress(operator []byte) (output []byte, err types2.Error) {
	store := u.Store()
	output, er := store.GetFishermanOutputAddress(operator)
	if er != nil {
		return nil, types2.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
