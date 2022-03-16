package utility

import (
	"math/big"
	"pocket/shared/crypto"
	"pocket/shared/modules"
	types2 "pocket/utility/types"
)

func (u *UtilityContext) HandleMessageStakeValidator(message *types2.MessageStakeValidator) types2.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types2.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	minStake, err := u.GetValidatorMinimumStake()
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
	// update account amount
	if err := u.SetAccount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(types2.ValidatorStakePoolName, amount); err != nil {
		return err
	}
	// ensure Validator doesn't already exist
	exists, err := u.GetValidatorExists(publicKey.Address())
	if err != nil {
		return err
	}
	if exists {
		return types2.ErrAlreadyExists()
	}
	// insert the Validator structure
	if err := u.InsertValidator(publicKey.Address(), message.PublicKey, message.OutputAddress, message.ServiceURL, message.Amount); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeValidator(message *types2.MessageEditStakeValidator) types2.Error {
	exists, err := u.GetValidatorExists(message.Address)
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
	// update account amount
	if err := u.SetAccount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(types2.ValidatorStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the validator structure
	if err := u.UpdateValidator(message.Address, message.ServiceURL, message.AmountToAdd); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeValidator(message *types2.MessageUnstakeValidator) types2.Error {
	status, err := u.GetValidatorStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != types2.StakedStatus {
		return types2.ErrInvalidStatus(status, types2.StakedStatus)
	}
	unstakingHeight, err := u.CalculateValidatorUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetValidatorUnstakingHeightAndStatus(message.Address, unstakingHeight, types2.UnstakingStatus); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeValidatorsThatAreReady() types2.Error {
	ValidatorsReadyToUnstake, err := u.GetValidatorsReadyToUnstake()
	if err != nil {
		return err
	}
	for _, validator := range ValidatorsReadyToUnstake {
		if err := u.SubPoolAmount(types2.ValidatorStakePoolName, validator.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.AddAccountAmountString(validator.GetOutputAddress(), validator.GetStakeAmount()); err != nil {
			return err
		}
		if err := u.DeleteValidator(validator.GetAddress()); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPausedValidators() types2.Error {
	maxPausedBlocks, err := u.GetValidatorMaxPausedBlocks()
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
	if err := u.UnstakeValidatorsPausedBefore(beforeHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessagePauseValidator(message *types2.MessagePauseValidator) types2.Error {
	height, err := u.GetValidatorPauseHeightIfExists(message.Address)
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
	if err := u.SetValidatorPauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnpauseValidator(message *types2.MessageUnpauseValidator) types2.Error {
	pausedHeight, err := u.GetValidatorPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == 0 {
		return types2.ErrNotPaused()
	}
	minPauseBlocks, err := u.GetValidatorMinimumPauseBlocks()
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
	if err := u.SetValidatorPauseHeight(message.Address, types2.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleByzantineValidators(lastBlockByzantineValidators [][]byte) types2.Error {
	latestBlockHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	maxMissedBlocks, err := u.GetValidatorMaxMissedBlocks()
	if err != nil {
		return err
	}
	for _, address := range lastBlockByzantineValidators {
		numberOfMissedBlocks, err := u.GetValidatorMissedBlocks(address)
		if err != nil {
			return err
		}
		if numberOfMissedBlocks > maxMissedBlocks {
			// pause the validator and reset missed blocks
			if err := u.SetValidatorPauseHeightAndMissedBlocks(address, latestBlockHeight, types2.ZeroInt); err != nil {
				return err
			}
			// burn validator for missing blocks
			burnPercentage, err := u.GetMissedBlocksBurnPercentage()
			if err != nil {
				return err
			}
			if err := u.BurnValidator(address, burnPercentage); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *UtilityContext) HandleProposal(proposer []byte) types2.Error {
	feesAndRewardsCollected, err := u.GetPoolAmount(types2.FeePoolName)
	if err != nil {
		return err
	}
	if err := u.SetPoolAmount(types2.FeePoolName, big.Int{}); err != nil {
		return err
	}
	proposerCutPercentage, err := u.GetProposerPercentageOfFees()
	if err != nil {
		return err
	}
	daoCutPercentage := 100 - proposerCutPercentage
	if daoCutPercentage < 0 {
		return types2.ErrInvalidProposerCutPercentage()
	}
	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feesAndRewardsCollected)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	amountToProposer, _ := feesAndRewardsCollectedFloat.Int(nil)
	amountToDAO := feesAndRewardsCollected.Sub(feesAndRewardsCollected, amountToProposer)
	if err := u.AddAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err := u.AddPoolAmount(types2.DAOPoolName, amountToDAO); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageDoubleSign(message *types2.MessageDoubleSign) types2.Error {
	evidenceAge := u.LatestHeight - message.VoteA.Height
	maxEvidenceAge, err := u.GetMaxEvidenceAgeInBlocks()
	if err != nil {
		return err
	}
	if evidenceAge > int64(maxEvidenceAge) {
		return types2.ErrMaxEvidenceAge()
	}
	pk, er := crypto.NewPublicKeyFromBytes(message.VoteB.PublicKey)
	if er != nil {
		return types2.ErrNewPublicKeyFromBytes(er)
	}
	doubleSigner := pk.Address()
	// burn validator for missing blocks
	burnPercentage, err := u.GetDoubleSignBurnPercentage()
	if err != nil {
		return err
	}
	if err := u.BurnValidator(doubleSigner, burnPercentage); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) BurnValidator(address []byte, percentage int) types2.Error {
	tokens, err := u.GetValidatorStakedTokens(address)
	if err != nil {
		return err
	}
	zeroBigInt := big.Int{}
	tokensFloat := new(big.Float).SetInt(tokens)
	tokensFloat.Mul(tokensFloat, big.NewFloat(float64(percentage)))
	tokensFloat.Quo(tokensFloat, big.NewFloat(100))
	truncatedTokens, _ := tokensFloat.Int(nil)
	if truncatedTokens.Cmp(zeroBigInt) == -1 {
		truncatedTokens = zeroBigInt
	}
	// remove from pool
	if err := u.SubPoolAmount(types2.ValidatorStakePoolName, BigIntToString(truncatedTokens)); err != nil {
		return err
	}
	// remove from validator
	if err := u.SetValidatorStakedTokens(address, truncatedTokens); err != nil {
		return err
	}
	// check to see if they fell below minimum stake
	minStake, err := u.GetValidatorMinimumStake()
	if err != nil {
		return err
	}
	// fell below minimum stake
	if minStake.Cmp(truncatedTokens) == 1 {
		unstakingHeight, err := u.CalculateValidatorUnstakingHeight()
		if err != nil {
			return err
		}
		if err := u.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, types2.UnstakingStatus); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) GetValidatorExists(address []byte) (exists bool, err types2.Error) {
	store := u.Store()
	exists, er := store.GetValidatorExists(address)
	if er != nil {
		return false, types2.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertValidator(address, publicKey, output []byte, serviceURL, amount string) types2.Error {
	store := u.Store()
	err := store.InsertValidator(address, publicKey, output, false, types2.StakedStatus, serviceURL, amount, types2.ZeroInt, types2.ZeroInt)
	if err != nil {
		return types2.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) UpdateValidator(address []byte, serviceURL, amount string) types2.Error {
	store := u.Store()
	err := store.UpdateValidator(address, serviceURL, amount)
	if err != nil {
		return types2.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteValidator(address []byte) types2.Error {
	store := u.Store()
	if err := store.DeleteValidator(address); err != nil {
		return types2.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetValidatorsReadyToUnstake() (Validators []modules.UnstakingActor, err types2.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingValidators, er := store.GetValidatorsReadyToUnstake(latestHeight, types2.UnstakingStatus)
	if er != nil {
		return nil, types2.ErrGetReadyToUnstake(er)
	}
	return unstakingValidators, nil
}

func (u *UtilityContext) UnstakeValidatorsPausedBefore(pausedBeforeHeight int64) types2.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateValidatorUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, types2.UnstakingStatus)
	if er != nil {
		return types2.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetValidatorStatus(address []byte) (status int, err types2.Error) {
	store := u.Store()
	status, er := store.GetValidatorStatus(address)
	if er != nil {
		return types2.ZeroInt, types2.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64, status int) (err types2.Error) {
	store := u.Store()
	if er := store.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, status); er != nil {
		return types2.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetValidatorPauseHeightIfExists(address []byte) (ValidatorPauseHeight int64, err types2.Error) {
	store := u.Store()
	ValidatorPauseHeight, er := store.GetValidatorPauseHeightIfExists(address)
	if er != nil {
		return types2.ZeroInt, types2.ErrGetPauseHeight(er)
	}
	return ValidatorPauseHeight, nil
}

func (u *UtilityContext) SetValidatorPauseHeight(address []byte, height int64) types2.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeight(address, height); err != nil {
		return types2.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateValidatorUnstakingHeight() (unstakingHeight int64, err types2.Error) {
	unstakingBlocks, err := u.GetValidatorUnstakingBlocks()
	if err != nil {
		return types2.ZeroInt, err
	}
	unstakingHeight, err = u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return types2.ZeroInt, err
	}
	return
}

func (u *UtilityContext) GetValidatorMissedBlocks(address []byte) (missedBlocks int, err types2.Error) {
	store := u.Store()
	missedBlocks, er := store.GetValidatorMissedBlocks(address)
	if er != nil {
		return types2.ZeroInt, types2.ErrGetMissedBlocks(err)
	}
	return missedBlocks, nil
}

func (u *UtilityContext) GetValidatorStakedTokens(address []byte) (tokens *big.Int, err types2.Error) {
	store := u.Store()
	validatorStakedTokens, er := store.GetValidatorStakedTokens(address)
	if er != nil {
		return nil, types2.ErrGetValidatorStakedTokens(err)
	}
	i, err := StringToBigInt(validatorStakedTokens)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (u *UtilityContext) SetValidatorStakedTokens(address []byte, tokens *big.Int) (err types2.Error) {
	store := u.Store()
	er := store.SetValidatorStakedTokens(address, BigIntToString(tokens))
	if er != nil {
		return types2.ErrSetValidatorStakedTokens(err)
	}
	return nil
}

func (u *UtilityContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) types2.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeightAndMissedBlocks(address, pauseHeight, missedBlocks); err != nil {
		return types2.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) GetMessageStakeValidatorSignerCandidates(msg *types2.MessageStakeValidator) (candidates [][]byte, err types2.Error) {
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types2.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, pk.Address())
	return
}

func (u *UtilityContext) GetMessageEditStakeValidatorSignerCandidates(msg *types2.MessageEditStakeValidator) (candidates [][]byte, err types2.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnstakeValidatorSignerCandidates(msg *types2.MessageUnstakeValidator) (candidates [][]byte, err types2.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageUnpauseValidatorSignerCandidates(msg *types2.MessageUnpauseValidator) (candidates [][]byte, err types2.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessagePauseValidatorSignerCandidates(msg *types2.MessagePauseValidator) (candidates [][]byte, err types2.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return
}

func (u *UtilityContext) GetMessageDoubleSignSignerCandidates(msg *types2.MessageDoubleSign) (candidates [][]byte, err types2.Error) {
	return [][]byte{msg.ReporterAddress}, nil
}

func (u *UtilityContext) GetValidatorOutputAddress(operator []byte) (output []byte, err types2.Error) {
	store := u.Store()
	output, er := store.GetValidatorOutputAddress(operator)
	if er != nil {
		return nil, types2.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}
