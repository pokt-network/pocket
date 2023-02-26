package utility

// Internal business logic specific to validator behaviour and interactions.

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// handleByzantineValidators identifies & handles byzantine or faulty validators.
// This includes validators who double signed, didn't sign at all or disagree with 2/3+ majority.
// IMPROVE: Need to add more logging to this function.
// INCOMPLETE: handleByzantineValidators is a WIP and needs to be fully designed, implemented, tested and documented
func (u *utilityContext) handleByzantineValidators(prevBlockByzantineValidators [][]byte) typesUtil.Error {
	maxMissedBlocks, err := u.getValidatorMaxMissedBlocks()
	if err != nil {
		return err
	}

	for _, address := range prevBlockByzantineValidators {
		// Get the latest number of missed blocks by the validator
		numMissedBlocks, err := u.store.GetValidatorMissedBlocks(address, u.height)
		if err != nil {
			return typesUtil.ErrGetMissedBlocks(err)
		}

		// increment missed blocks
		numMissedBlocks++

		// handle if under the threshold of max missed blocks
		if numMissedBlocks < maxMissedBlocks {
			if err := u.store.SetValidatorMissedBlocks(address, numMissedBlocks); err != nil {
				return typesUtil.ErrSetMissedBlocks(err)
			}
			continue
		}

		// pause the validator for exceeding the threshold: numMissedBlocks >= maxMissedBlocks
		if err := u.store.SetValidatorPauseHeight(address, u.height); err != nil {
			return typesUtil.ErrSetPauseHeight(err)
		}
		// update the number of blocks it missed
		if err := u.store.SetValidatorMissedBlocks(address, numMissedBlocks); err != nil {
			return typesUtil.ErrSetMissedBlocks(err)
		}
		// burn validator for missing blocks
		if err := u.burnValidator(address); err != nil {
			return err
		}
	}
	return nil
}

// burnValidator burns a validator's stake based on governance parameters for missing blocks
// and begins unstaking if the stake falls below the necessary threshold
// REFACTOR: Extend this to support burning other actors types & pools once the logic is implemented
// ADDTEST: There are no good tests for this functionality, which MUST be added.
func (u *utilityContext) burnValidator(addr []byte) typesUtil.Error {
	actorType := coreTypes.ActorType_ACTOR_TYPE_VAL
	actorPool := coreTypes.Pools_POOLS_VALIDATOR_STAKE

	stakeAmount, err := u.getActorStakeAmount(actorType, addr)
	if err != nil {
		return err
	}

	burnPercent, err := u.getMissedBlocksBurnPercentage()
	if err != nil {
		return err
	}

	// burnAmount = currentStake * burnPercent / 100
	burnAmount := new(big.Float).SetInt(stakeAmount)
	burnAmount.Mul(burnAmount, big.NewFloat(float64(burnPercent)))
	burnAmount.Quo(burnAmount, big.NewFloat(100))
	burnAmountTruncated, _ := burnAmount.Int(nil)

	// Round up to 0 if -ve
	zeroBigInt := big.NewInt(0)
	if burnAmountTruncated.Cmp(zeroBigInt) == -1 {
		burnAmountTruncated = zeroBigInt
	}

	// remove burnt stake amount from the pool
	if err := u.subPoolAmount(actorPool.FriendlyName(), burnAmountTruncated); err != nil {
		return err
	}

	// remove burnt stake from the actor
	newAmountAfterBurn := big.NewInt(0).Sub(stakeAmount, burnAmountTruncated)
	if err := u.setActorStakeAmount(actorType, addr, newAmountAfterBurn); err != nil {
		return err
	}

	// Need to check if the actor needs to be unstaked
	minRequiredStake, err := u.getValidatorMinimumStake()
	if err != nil {
		return err
	}

	// Check if amount after burn is below the min required stake
	if minRequiredStake.Cmp(newAmountAfterBurn) == -1 {
		unbondingHeight, err := u.getUnbondingHeight(actorType)
		if err != nil {
			return err
		}
		if err := u.setActorUnbondingHeight(actorType, addr, unbondingHeight); err != nil {
			return err
		}
	}

	return nil
}
