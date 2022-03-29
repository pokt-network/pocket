package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	utilTypes "github.com/pokt-network/pocket/utility/types"
)

func (u *UtilityContext) HandleMessageStakeValidator(message *utilTypes.MessageStakeValidator) types.Error {
	publicKey, er := crypto.NewPublicKeyFromBytes(message.PublicKey)
	if er != nil {
		return types.ErrNewPublicKeyFromBytes(er)
	}
	// ensure above minimum stake
	minStake, err := u.GetValidatorMinimumStake()
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
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(utilTypes.ValidatorStakePoolName, amount); err != nil {
		return err
	}
	// ensure Validator doesn't already exist
	exists, err := u.GetValidatorExists(publicKey.Address())
	if err != nil {
		return err
	}
	if exists {
		return types.ErrAlreadyExists()
	}
	// insert the Validator structure
	if err := u.InsertValidator(publicKey.Address(), message.PublicKey, message.OutputAddress, message.ServiceUrl, message.Amount); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageEditStakeValidator(message *utilTypes.MessageEditStakeValidator) types.Error {
	exists, err := u.GetValidatorExists(message.Address)
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
	// update account amount
	if err := u.SetAccountAmount(message.Signer, signerAccountAmount); err != nil {
		return err
	}
	// move funds from account to pool
	if err := u.AddPoolAmount(utilTypes.ValidatorStakePoolName, amountToAdd); err != nil {
		return err
	}
	// insert the validator structure
	if err := u.UpdateValidator(message.Address, message.ServiceUrl, message.AmountToAdd); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnstakeValidator(message *utilTypes.MessageUnstakeValidator) types.Error {
	status, err := u.GetValidatorStatus(message.Address)
	if err != nil {
		return err
	}
	// validate is staked
	if status != utilTypes.StakedStatus {
		return types.ErrInvalidStatus(status, utilTypes.StakedStatus)
	}
	unstakingHeight, err := u.CalculateValidatorUnstakingHeight()
	if err != nil {
		return err
	}
	if err := u.SetValidatorUnstakingHeightAndStatus(message.Address, unstakingHeight); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) UnstakeValidatorsThatAreReady() types.Error {
	validatorsReadyToUnstake, err := u.GetValidatorsReadyToUnstake()
	if err != nil {
		return err
	}
	// If unstaking even a single validator fails, the whole operation falls through.
	for _, validator := range validatorsReadyToUnstake {
		if err := u.SubPoolAmount(utilTypes.ValidatorStakePoolName, validator.GetStakeAmount()); err != nil {
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

func (u *UtilityContext) BeginUnstakingMaxPausedValidators() types.Error {
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

func (u *UtilityContext) HandleMessagePauseValidator(message *utilTypes.MessagePauseValidator) types.Error {
	height, err := u.GetValidatorPauseHeightIfExists(message.Address)
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
	if err := u.SetValidatorPauseHeight(message.Address, height); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageUnpauseValidator(message *utilTypes.MessageUnpauseValidator) types.Error {
	pausedHeight, err := u.GetValidatorPauseHeightIfExists(message.Address)
	if err != nil {
		return err
	}
	if pausedHeight == 0 {
		return types.ErrNotPaused()
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
		return types.ErrNotReadyToUnpause()
	}
	if err := u.SetValidatorPauseHeight(message.Address, utilTypes.ZeroInt); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleByzantineValidators(lastBlockByzantineValidators [][]byte) types.Error {
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
		// increment missed blocks
		numberOfMissedBlocks++
		// handle if over the threshold
		if numberOfMissedBlocks >= maxMissedBlocks {
			// pause the validator and reset missed blocks
			if err := u.SetValidatorPauseHeightAndMissedBlocks(address, latestBlockHeight, utilTypes.ZeroInt); err != nil {
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
		} else if err := u.SetValidatorMissedBlocks(address, numberOfMissedBlocks); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) HandleProposalRewards(proposer []byte) types.Error {
	feesAndRewardsCollected, err := u.GetPoolAmount(utilTypes.FeePoolName)
	if err != nil {
		return err
	}
	if err := u.SetPoolAmount(utilTypes.FeePoolName, big.NewInt(0)); err != nil {
		return err
	}
	proposerCutPercentage, err := u.GetProposerPercentageOfFees()
	if err != nil {
		return err
	}
	daoCutPercentage := 100 - proposerCutPercentage
	if daoCutPercentage < 0 || daoCutPercentage > 100 {
		return types.ErrInvalidProposerCutPercentage()
	}
	amountToProposerFloat := new(big.Float).SetInt(feesAndRewardsCollected)
	amountToProposerFloat.Mul(amountToProposerFloat, big.NewFloat(float64(proposerCutPercentage)))
	amountToProposerFloat.Quo(amountToProposerFloat, big.NewFloat(100))
	amountToProposer, _ := amountToProposerFloat.Int(nil)
	amountToDAO := feesAndRewardsCollected.Sub(feesAndRewardsCollected, amountToProposer)
	if err := u.AddAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err := u.AddPoolAmount(utilTypes.DAOPoolName, amountToDAO); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) HandleMessageDoubleSign(message *utilTypes.MessageDoubleSign) types.Error {
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
	if err := u.BurnValidator(doubleSigner, burnPercentage); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) BurnValidator(address []byte, percentage int) types.Error {
	tokens, err := u.GetValidatorStakedTokens(address)
	if err != nil {
		return err
	}
	zeroBigInt := big.NewInt(0)
	tokensFloat := new(big.Float).SetInt(tokens)
	tokensFloat.Mul(tokensFloat, big.NewFloat(float64(percentage)))
	tokensFloat.Quo(tokensFloat, big.NewFloat(100))
	truncatedTokens, _ := tokensFloat.Int(nil)
	if truncatedTokens.Cmp(zeroBigInt) == -1 {
		truncatedTokens = zeroBigInt
	}
	newTokensAfterBurn := big.NewInt(0).Sub(tokens, truncatedTokens)
	// remove from pool
	if err := u.SubPoolAmount(utilTypes.ValidatorStakePoolName, types.BigIntToString(truncatedTokens)); err != nil {
		return err
	}
	// remove from validator
	if err := u.SetValidatorStakedTokens(address, newTokensAfterBurn); err != nil {
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
		if err := u.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) GetValidatorExists(address []byte) (bool, types.Error) {
	store := u.Store()
	exists, er := store.GetValidatorExists(address)
	if er != nil {
		return false, types.ErrGetExists(er)
	}
	return exists, nil
}

func (u *UtilityContext) InsertValidator(address, publicKey, output []byte, serviceURL, amount string) types.Error {
	store := u.Store()
	err := store.InsertValidator(address, publicKey, output, false, utilTypes.StakedStatus, serviceURL, amount, utilTypes.ZeroInt, utilTypes.ZeroInt)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) UpdateValidator(address []byte, serviceURL, amount string) types.Error {
	store := u.Store()
	err := store.UpdateValidator(address, serviceURL, amount)
	if err != nil {
		return types.ErrInsert(err)
	}
	return nil
}

func (u *UtilityContext) DeleteValidator(address []byte) types.Error {
	store := u.Store()
	if err := store.DeleteValidator(address); err != nil {
		return types.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) GetValidatorsReadyToUnstake() ([]*types.UnstakingActor, types.Error) {
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return nil, err
	}
	unstakingValidators, er := store.GetValidatorsReadyToUnstake(latestHeight, utilTypes.UnstakingStatus)
	if er != nil {
		return nil, types.ErrGetReadyToUnstake(er)
	}
	return unstakingValidators, nil
}

func (u *UtilityContext) UnstakeValidatorsPausedBefore(pausedBeforeHeight int64) types.Error {
	store := u.Store()
	unstakingHeight, err := u.CalculateValidatorUnstakingHeight()
	if err != nil {
		return err
	}
	er := store.SetValidatorsStatusAndUnstakingHeightPausedBefore(pausedBeforeHeight, unstakingHeight, utilTypes.UnstakingStatus)
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) GetValidatorStatus(address []byte) (int, types.Error) {
	store := u.Store()
	status, er := store.GetValidatorStatus(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) types.Error {
	store := u.Store()
	er := store.SetValidatorMissedBlocks(address, missedBlocks)
	if er != nil {
		return types.ErrSetMissedBlocks(er)
	}
	return nil
}

func (u *UtilityContext) SetValidatorUnstakingHeightAndStatus(address []byte, unstakingHeight int64) types.Error {
	store := u.Store()
	if er := store.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, utilTypes.UnstakingStatus); er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) GetValidatorPauseHeightIfExists(address []byte) (int64, types.Error) {
	store := u.Store()
	ValidatorPauseHeight, er := store.GetValidatorPauseHeightIfExists(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetPauseHeight(er)
	}
	return ValidatorPauseHeight, nil
}

func (u *UtilityContext) SetValidatorPauseHeight(address []byte, height int64) types.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeight(address, height); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) CalculateValidatorUnstakingHeight() (int64, types.Error) {
	unstakingBlocks, err := u.GetValidatorUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	unstakingHeight, err := u.CalculateUnstakingHeight(unstakingBlocks)
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	return unstakingHeight, nil
}

func (u *UtilityContext) GetValidatorMissedBlocks(address []byte) (int, types.Error) {
	store := u.Store()
	missedBlocks, er := store.GetValidatorMissedBlocks(address)
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetMissedBlocks(er)
	}
	return missedBlocks, nil
}

func (u *UtilityContext) GetValidatorStakedTokens(address []byte) (*big.Int, types.Error) {
	store := u.Store()
	validatorStakedTokens, er := store.GetValidatorStakedTokens(address)
	if er != nil {
		return nil, types.ErrGetValidatorStakedTokens(er)
	}
	i, err := types.StringToBigInt(validatorStakedTokens)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (u *UtilityContext) SetValidatorStakedTokens(address []byte, tokens *big.Int) types.Error {
	store := u.Store()
	er := store.SetValidatorStakedTokens(address, types.BigIntToString(tokens))
	if er != nil {
		return types.ErrSetValidatorStakedTokens(er)
	}
	return nil
}

func (u *UtilityContext) SetValidatorPauseHeightAndMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) types.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeightAndMissedBlocks(address, pauseHeight, missedBlocks); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) GetMessageStakeValidatorSignerCandidates(msg *utilTypes.MessageStakeValidator) ([][]byte, types.Error) {
	candidates := make([][]byte, 0)
	candidates = append(candidates, msg.OutputAddress)
	pk, er := crypto.NewPublicKeyFromBytes(msg.PublicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	candidates = append(candidates, pk.Address())
	return candidates, nil
}

func (u *UtilityContext) GetMessageEditStakeValidatorSignerCandidates(msg *utilTypes.MessageEditStakeValidator) ([][]byte, types.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnstakeValidatorSignerCandidates(msg *utilTypes.MessageUnstakeValidator) ([][]byte, types.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageUnpauseValidatorSignerCandidates(msg *utilTypes.MessageUnpauseValidator) ([][]byte, types.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessagePauseValidatorSignerCandidates(msg *utilTypes.MessagePauseValidator) ([][]byte, types.Error) {
	output, err := u.GetValidatorOutputAddress(msg.Address)
	if err != nil {
		return nil, err
	}
	candidates := make([][]byte, 0)
	candidates = append(candidates, output)
	candidates = append(candidates, msg.Address)
	return candidates, nil
}

func (u *UtilityContext) GetMessageDoubleSignSignerCandidates(msg *utilTypes.MessageDoubleSign) ([][]byte, types.Error) {
	return [][]byte{msg.ReporterAddress}, nil
}

func (u *UtilityContext) GetValidatorOutputAddress(operator []byte) ([]byte, types.Error) {
	store := u.Store()
	output, er := store.GetValidatorOutputAddress(operator)
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}

func (u *UtilityContext) CalculateUnstakingHeight(unstakingBlocks int64) (int64, types.Error) {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return utilTypes.ZeroInt, err
	}
	return unstakingBlocks + latestHeight, nil
}
