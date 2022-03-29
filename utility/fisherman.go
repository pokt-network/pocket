package utility

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageTestScore(message *typesUtil.MessageTestScore) types.Error {
	panic("TODO")
}

func (u *UtilityContext) HandleMessageProveTestScore(message *typesUtil.MessageProveTestScore) types.Error {
	panic("TODO")
}

func (u *UtilityContext) HandleMessageStakeFisherman(message *typesUtil.MessageStakeFisherman) types.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	minStake, err := u.GetFishermanMinimumStake()
	if err != nil {
		return err
	}
	amount, err := types.StringToBigInt(message.Amount)
	if err != nil {
		return err
	}
	if types.BigIntLessThan(amount, minStake) {
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
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(typesUtil.FishermanStakePoolName, amount); err != nil {
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
	if err := u.InsertFisherman(publicKey.Address(), message.PublicKey, message.OutputAddress, message.ServiceUrl, message.Amount, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeFisherman(message *typesUtil.MessageEditStakeFisherman) types.Error {
	exists, err := u.GetFishermanExists(message.Address)
	if err != nil {
		return err
	}
	if !exists {
		return types.ErrNotExists()
	}
	amountToAdd, err := types.StringToBigInt(message.AmountToAdd)
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
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(typesUtil.FishermanStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the Fisherman structure
	if err := u.UpdateFisherman(message.Address, message.ServiceUrl, message.AmountToAdd, message.Chains); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeFisherman(message *typesUtil.MessageUnstakeFisherman) types.Error {
	status, err := u.GetFishermanStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != typesUtil.StakedStatus {
		return types.ErrInvalidStatus(status, typesUtil.StakedStatus)
	}
	unstakingHeight, err := u.CalculateFishermanUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetFishermanUnstakingHeightAndStatus(message.Address, unstakingHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeFishermenThatAreReady() types.Error {
	fishermansReadyToUnstake, err := u.GetFishermenReadyToUnstake()
	if err != nil {
		return err
	}
	for _, fisherman := range fishermansReadyToUnstake {
		if err := u.SubPoolAmount(typesUtil.FishermanStakePoolName, fisherman.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.AddAccountAmountString(fisherman.GetOutputAddress(), fisherman.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.DeleteFisherman(fisherman.GetAddress()); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPausedFishermen() types.Error {
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
	if err := u.UnstakeFishermenPausedBefore(beforeHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessagePauseFisherman(message *typesUtil.MessagePauseFisherman) types.Error {
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

func (u *UtilityContext) HandleMessageFishermanPauseServiceNode(message *typesUtil.MessageFishermanPauseServiceNode) types.Error {
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

func (u *UtilityContext) HandleMessageUnpauseFisherman(message *typesUtil.MessageUnpauseFisherman) types.Error {
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
	if err := u.SetFishermanPauseHeight(message.Address, typesUtil.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) GetFishermanExists(address []byte) (bool, types.Error) {
	store := u.Store()
	exists, er := store.GetFishermanExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertFisherman(address, publicKey, output []byte, serviceURL, amount string, chains []string) types.Error {
	store := u.Store()
	err := store.InsertFisherman(address, publicKey, output, false, typesUtil.StakedStatus, serviceURL, amount, chains, typesUtil.ZeroInt, typesUtil.ZeroInt)
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

func (u *UtilityContext) GetFishermenReadyToUnstake() ([]*types.UnstakingActor, types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingFishermans, er := store.GetFishermanReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
	if er != nil {
		return nil, types.ErrGetReadyToUnstake(er)
	}
	return unstakingFishermans, nil
}

func (u *UtilityContext) UnstakeFishermenPausedBefore(pausedBeforeHeight int64) types.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateFishermanUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetFishermansStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetFishermanStatus(address []byte) (int, types.Error) {
	store := u.Store()
	status, er := store.GetFishermanStatus(address)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetFishermanUnstakingHeightAndStatus(address []byte, unstakingHeight int64) types.Error {
	store := u.Store()
	if er := store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus); er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetFishermanPauseHeightIfExists(address []byte) (int64, types.Error) {
	store := u.Store()
	fishermanPauseHeight, er := store.GetFishermanPauseHeightIfExists(address)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetPauseHeight(er)
	}
	return fishermanPauseHeight, nil
}

func (u *UtilityContext) SetFishermanPauseHeight(address []byte, height int64) types.Error {
	store := u.Store()
	if err := store.SetFishermanPauseHeight(address, height); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateFishermanUnstakingHeight() (int64, types.Error) {
	unstakingBlocks, err := u.GetFishermanUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	unstakingHeight, err := u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	return unstakingHeight, nil
}

func (u *UtilityContext) GetMessageStakeFishermanSignerCandidates(msg *typesUtil.MessageStakeFisherman) ([][]byte, types.Error) {
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, msg.OutputAddress)
	candidates = append(candidates, pk.Address())
	return candidates, nil
}

func (u *UtilityContext) GetMessageEditStakeFishermanSignerCandidates(msg *typesUtil.MessageEditStakeFisherman) ([][]byte, types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnstakeFishermanSignerCandidates(msg *typesUtil.MessageUnstakeFisherman) ([][]byte, types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnpauseFishermanSignerCandidates(msg *typesUtil.MessageUnpauseFisherman) ([][]byte, types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessagePauseFishermanSignerCandidates(msg *typesUtil.MessagePauseFisherman) ([][]byte, types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeSignerCandidates(msg *typesUtil.MessageFishermanPauseServiceNode) ([][]byte, types.Error) {
	output, err := u.GetFishermanOutputAddress(msg.Reporter)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Reporter)
	return candidates, nil
}

func (u *UtilityContext) GetFishermanOutputAddress(operator []byte) ([]byte, types.Error) {
	store := u.Store()
	output, er := store.GetFishermanOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
