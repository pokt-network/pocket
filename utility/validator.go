package utility

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// ADDTEST: There are no good tests for this functionality, which MUST be added.
func (u *utilityContext) burnValidator(burnPercent int, addr []byte) typesUtil.Error {
	// TODO: Will need to extend this to support burning from other actors types & pools when the logic is implemented
	validatorActorType := coreTypes.ActorType_ACTOR_TYPE_VAL
	validatorPool := coreTypes.Pools_POOLS_VALIDATOR_STAKE

	stakeAmount, err := u.getActorStakeAmount(validatorActorType, addr)
	if err != nil {
		return err
	}

	// stake after burn = current take * newStake = currentStake * burnPercent / 100
	burnAmount := new(big.Float).SetInt(stakeAmount)
	burnAmount.Mul(burnAmount, big.NewFloat(float64(burnPercent)))
	burnAmount.Quo(burnAmount, big.NewFloat(100))
	burnAmountTruncated, _ := burnAmount.Int(nil)

	// Round up to 0 if -ve
	zeroBigInt := big.NewInt(0)
	if burnAmountTruncated.Cmp(zeroBigInt) == -1 {
		burnAmountTruncated = zeroBigInt
	}

	newAmountAfterBurn := big.NewInt(0).Sub(stakeAmount, burnAmountTruncated)

	// remove from pool
	if err := u.subPoolAmount(validatorPool.FriendlyName(), burnAmountTruncated); err != nil {
		return err
	}

	// remove from actor
	if err := u.setActorStakeAmount(validatorActorType, addr, newAmountAfterBurn); err != nil {
		return err
	}

	// Need to check if new stake is below min required stake
	minStake, err := u.getValidatorMinimumStake()
	if err != nil {
		return err
	}

	// Check if amount after burn is below the min required stake
	if minStake.Cmp(newAmountAfterBurn) == -1 {
		unstakingHeight, err := u.getUnbondingHeight(validatorActorType)
		if err != nil {
			return err
		}
		if err := u.setActorUnstakingHeight(validatorActorType, addr, unstakingHeight); err != nil {
			return err
		}
	}

	return nil
}
