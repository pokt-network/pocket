package converters

import (
	typesPers "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

func toPersistenceActor(actor modules.Actor) *typesPers.Actor {
	return &typesPers.Actor{
		Address:      actor.GetAddress(),
		PublicKey:    actor.GetPublicKey(),
		StakedAmount: actor.GetStakedAmount(),
		GenericParam: actor.GetGenericParam(),
	}
}

func ToPersistenceActors(actors []modules.Actor) []*typesPers.Actor {
	r := make([]*typesPers.Actor, 0)
	for _, a := range actors {
		r = append(r, toPersistenceActor(a))
	}
	return r
}

func toPersistenceAccount(account modules.Account) *typesPers.Account {
	return &typesPers.Account{
		Address: account.GetAddress(),
		Amount:  account.GetAmount(),
	}
}

func ToPersistenceAccounts(accounts []modules.Account) []*typesPers.Account {
	r := make([]*typesPers.Account, 0)
	for _, a := range accounts {
		r = append(r, toPersistenceAccount(a))
	}
	return r
}

func ToPersistenceParams(params modules.Params) *typesPers.Params {
	return &typesPers.Params{
		BlocksPerSession:                         params.GetBlocksPerSession(),
		AppMinimumStake:                          params.GetAppMinimumStake(),
		AppMaxChains:                             params.GetAppMaxChains(),
		AppBaselineStakeRate:                     params.GetAppBaselineStakeRate(),
		AppStakingAdjustment:                     params.GetAppStakingAdjustment(),
		AppUnstakingBlocks:                       params.GetAppUnstakingBlocks(),
		AppMinimumPauseBlocks:                    params.GetAppMinimumPauseBlocks(),
		AppMaxPauseBlocks:                        params.GetAppMaxPauseBlocks(),
		ServiceNodeMinimumStake:                  params.GetServiceNodeMinimumStake(),
		ServiceNodeMaxChains:                     params.GetServiceNodeMaxChains(),
		ServiceNodeUnstakingBlocks:               params.GetServiceNodeUnstakingBlocks(),
		ServiceNodeMinimumPauseBlocks:            params.GetServiceNodeMinimumPauseBlocks(),
		ServiceNodeMaxPauseBlocks:                params.GetServiceNodeMaxPauseBlocks(),
		ServiceNodesPerSession:                   params.GetServiceNodesPerSession(),
		FishermanMinimumStake:                    params.GetFishermanMinimumStake(),
		FishermanMaxChains:                       params.GetFishermanMaxChains(),
		FishermanUnstakingBlocks:                 params.GetFishermanUnstakingBlocks(),
		FishermanMinimumPauseBlocks:              params.GetFishermanMinimumPauseBlocks(),
		FishermanMaxPauseBlocks:                  params.GetFishermanMaxPauseBlocks(),
		ValidatorMinimumStake:                    params.GetValidatorMinimumStake(),
		ValidatorUnstakingBlocks:                 params.GetValidatorUnstakingBlocks(),
		ValidatorMinimumPauseBlocks:              params.GetValidatorMinimumPauseBlocks(),
		ValidatorMaxPauseBlocks:                  params.GetValidatorMaxPauseBlocks(),
		ValidatorMaximumMissedBlocks:             params.GetValidatorMaximumMissedBlocks(),
		ValidatorMaxEvidenceAgeInBlocks:          params.GetValidatorMaxEvidenceAgeInBlocks(),
		ProposerPercentageOfFees:                 params.GetProposerPercentageOfFees(),
		MissedBlocksBurnPercentage:               params.GetMissedBlocksBurnPercentage(),
		DoubleSignBurnPercentage:                 params.GetDoubleSignBurnPercentage(),
		MessageDoubleSignFee:                     params.GetMessageDoubleSignFee(),
		MessageSendFee:                           params.GetMessageSendFee(),
		MessageStakeFishermanFee:                 params.GetMessageStakeFishermanFee(),
		MessageEditStakeFishermanFee:             params.GetMessageEditStakeFishermanFee(),
		MessageUnstakeFishermanFee:               params.GetMessageUnstakeFishermanFee(),
		MessagePauseFishermanFee:                 params.GetMessagePauseFishermanFee(),
		MessageUnpauseFishermanFee:               params.GetMessageUnpauseFishermanFee(),
		MessageFishermanPauseServiceNodeFee:      params.GetMessageFishermanPauseServiceNodeFee(),
		MessageTestScoreFee:                      params.GetMessageTestScoreFee(),
		MessageProveTestScoreFee:                 params.GetMessageProveTestScoreFee(),
		MessageStakeAppFee:                       params.GetMessageStakeAppFee(),
		MessageEditStakeAppFee:                   params.GetMessageEditStakeAppFee(),
		MessageUnstakeAppFee:                     params.GetMessageUnstakeAppFee(),
		MessagePauseAppFee:                       params.GetMessagePauseAppFee(),
		MessageUnpauseAppFee:                     params.GetMessageUnpauseAppFee(),
		MessageStakeValidatorFee:                 params.GetMessageStakeValidatorFee(),
		MessageEditStakeValidatorFee:             params.GetMessageEditStakeValidatorFee(),
		MessageUnstakeValidatorFee:               params.GetMessageUnstakeValidatorFee(),
		MessagePauseValidatorFee:                 params.GetMessagePauseValidatorFee(),
		MessageUnpauseValidatorFee:               params.GetMessageUnpauseValidatorFee(),
		MessageStakeServiceNodeFee:               params.GetMessageStakeServiceNodeFee(),
		MessageEditStakeServiceNodeFee:           params.GetMessageEditStakeServiceNodeFee(),
		MessageUnstakeServiceNodeFee:             params.GetMessageUnstakeServiceNodeFee(),
		MessagePauseServiceNodeFee:               params.GetMessagePauseServiceNodeFee(),
		MessageUnpauseServiceNodeFee:             params.GetMessageUnpauseServiceNodeFee(),
		MessageChangeParameterFee:                params.GetMessageChangeParameterFee(),
		AclOwner:                                 params.GetAclOwner(),
		BlocksPerSessionOwner:                    params.GetBlocksPerSessionOwner(),
		AppMinimumStakeOwner:                     params.GetAppMinimumStakeOwner(),
		AppMaxChainsOwner:                        params.GetAppMaxChainsOwner(),
		AppBaselineStakeRateOwner:                params.GetAppBaselineStakeRateOwner(),
		AppStakingAdjustmentOwner:                params.GetAppStakingAdjustmentOwner(),
		AppUnstakingBlocksOwner:                  params.GetAppUnstakingBlocksOwner(),
		AppMinimumPauseBlocksOwner:               params.GetAppMinimumPauseBlocksOwner(),
		AppMaxPausedBlocksOwner:                  params.GetAppMaxPausedBlocksOwner(),
		ServiceNodeMinimumStakeOwner:             params.GetServiceNodeMinimumStakeOwner(),
		ServiceNodeMaxChainsOwner:                params.GetServiceNodeMaxChainsOwner(),
		ServiceNodeUnstakingBlocksOwner:          params.GetServiceNodeUnstakingBlocksOwner(),
		ServiceNodeMinimumPauseBlocksOwner:       params.GetServiceNodeMinimumPauseBlocksOwner(),
		ServiceNodeMaxPausedBlocksOwner:          params.GetServiceNodeMaxPausedBlocksOwner(),
		ServiceNodesPerSessionOwner:              params.GetServiceNodesPerSessionOwner(),
		FishermanMinimumStakeOwner:               params.GetFishermanMinimumStakeOwner(),
		FishermanMaxChainsOwner:                  params.GetFishermanMaxChainsOwner(),
		FishermanUnstakingBlocksOwner:            params.GetFishermanUnstakingBlocksOwner(),
		FishermanMinimumPauseBlocksOwner:         params.GetFishermanMinimumPauseBlocksOwner(),
		FishermanMaxPausedBlocksOwner:            params.GetFishermanMaxPausedBlocksOwner(),
		ValidatorMinimumStakeOwner:               params.GetValidatorMinimumStakeOwner(),
		ValidatorUnstakingBlocksOwner:            params.GetValidatorUnstakingBlocksOwner(),
		ValidatorMinimumPauseBlocksOwner:         params.GetValidatorMinimumPauseBlocksOwner(),
		ValidatorMaxPausedBlocksOwner:            params.GetValidatorMaxPausedBlocksOwner(),
		ValidatorMaximumMissedBlocksOwner:        params.GetValidatorMaximumMissedBlocksOwner(),
		ValidatorMaxEvidenceAgeInBlocksOwner:     params.GetValidatorMaxEvidenceAgeInBlocksOwner(),
		ProposerPercentageOfFeesOwner:            params.GetProposerPercentageOfFeesOwner(),
		MissedBlocksBurnPercentageOwner:          params.GetMissedBlocksBurnPercentageOwner(),
		DoubleSignBurnPercentageOwner:            params.GetDoubleSignBurnPercentageOwner(),
		MessageDoubleSignFeeOwner:                params.GetMessageDoubleSignFeeOwner(),
		MessageSendFeeOwner:                      params.GetMessageSendFeeOwner(),
		MessageStakeFishermanFeeOwner:            params.GetMessageStakeFishermanFeeOwner(),
		MessageEditStakeFishermanFeeOwner:        params.GetMessageEditStakeFishermanFeeOwner(),
		MessageUnstakeFishermanFeeOwner:          params.GetMessageUnstakeFishermanFeeOwner(),
		MessagePauseFishermanFeeOwner:            params.GetMessagePauseFishermanFeeOwner(),
		MessageUnpauseFishermanFeeOwner:          params.GetMessageUnpauseFishermanFeeOwner(),
		MessageFishermanPauseServiceNodeFeeOwner: params.GetMessageFishermanPauseServiceNodeFeeOwner(),
		MessageTestScoreFeeOwner:                 params.GetMessageTestScoreFeeOwner(),
		MessageProveTestScoreFeeOwner:            params.GetMessageProveTestScoreFeeOwner(),
		MessageStakeAppFeeOwner:                  params.GetMessageStakeAppFeeOwner(),
		MessageEditStakeAppFeeOwner:              params.GetMessageEditStakeAppFeeOwner(),
		MessageUnstakeAppFeeOwner:                params.GetMessageUnstakeAppFeeOwner(),
		MessagePauseAppFeeOwner:                  params.GetMessagePauseAppFeeOwner(),
		MessageUnpauseAppFeeOwner:                params.GetMessageUnpauseAppFeeOwner(),
		MessageStakeValidatorFeeOwner:            params.GetMessageStakeValidatorFeeOwner(),
		MessageEditStakeValidatorFeeOwner:        params.GetMessageEditStakeValidatorFeeOwner(),
		MessageUnstakeValidatorFeeOwner:          params.GetMessageUnstakeValidatorFeeOwner(),
		MessagePauseValidatorFeeOwner:            params.GetMessagePauseValidatorFeeOwner(),
		MessageUnpauseValidatorFeeOwner:          params.GetMessageUnpauseValidatorFeeOwner(),
		MessageStakeServiceNodeFeeOwner:          params.GetMessageStakeServiceNodeFeeOwner(),
		MessageEditStakeServiceNodeFeeOwner:      params.GetMessageEditStakeServiceNodeFeeOwner(),
		MessageUnstakeServiceNodeFeeOwner:        params.GetMessageUnstakeServiceNodeFeeOwner(),
		MessagePauseServiceNodeFeeOwner:          params.GetMessagePauseServiceNodeFeeOwner(),
		MessageUnpauseServiceNodeFeeOwner:        params.GetMessageUnpauseServiceNodeFeeOwner(),
		MessageChangeParameterFeeOwner:           params.GetMessageChangeParameterFeeOwner(),
	}
}
