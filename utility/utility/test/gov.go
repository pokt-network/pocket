package test

import (
	"math/big"
	"pocket/utility/shared/crypto"
	"pocket/utility/utility"
	"pocket/utility/utility/types"
)

var (
	// NOTE: this is for fun illustration purposes... The addresses begin with DA0, DA0, and FEE :)
	// Of course, in a production network the params / owners must be set in the genesis file
	DefaultParamsOwner, _          = crypto.NewPrivateKey("ff538589deb7f28bbce1ba68b37d2efc0eaa03204b36513cf88422a875559e38d6cbe0430ddd85a5e48e0c99ef3dea47bf0d1a83c6e6ad1640f72201dc8a0120")
	DefaultDAOPool, _              = crypto.NewPrivateKey("b1dfb25a67dadf9cdd39927b86166149727649af3a3143e66e558652f8031f3faacaa24a69bcf2819ed97ab5ed8d1e490041e5c7ef9e1eddba8b5678f997ae58")
	DefaultFeeCollector, _         = crypto.NewPrivateKey("bdc02826b5da77b90a5d1550443b3f007725cc654c10002aa01e65a131f3464b826f8e7911fa89b4bd6659c3175114d714c60bac63acc63817c0d3a4ed2fdab8")
	DefaultFishermanStakePool, _   = crypto.NewPrivateKey("f3dd5c8ccd9a7c8d0afd36424c6fbe8ead55315086ef3d0d03ce8c7357e5e306733a711adb6fc8fbef6a3e2ac2db7842433053a23c751d19573ab85b52316f67")
	DefaultServiceNodeStakePool, _ = crypto.NewPrivateKey("b4e4426ed014d5ee89949e6f60c406c328e4fce466cd25f4697a41046b34313097a8cc38033822da010422851062ae6b21b8e29d4c34193b7d8fa0f37b6593b6")
	DefaultValidatorStakePool, _   = crypto.NewPrivateKey("e0b8b7cdb33f11a8d70eb05070e53b02fe74f4499aed7b159bd2dd256e356d67664b5b682e40ee218e5feea05c2a1bb595ec15f3850c92b571cdf950b4d9ba23")
	DefaultAppStakePool, _         = crypto.NewPrivateKey("429627bac8dc322f0aeeb2b8f25b329899b7ebb9605d603b5fb74557b13357e50834e9575c19d9d7d664ec460a98abb2435ece93440eb482c87d5b7259a8d271")
)

func DefaultParams() *Params {
	return &Params{
		BlocksPerSession:                         4,
		AppMinimumStake:                          types.BigIntToString(big.NewInt(15000000000)),
		AppMaxChains:                             15,
		AppBaselineStakeRate:                     100,
		AppStakingAdjustment:                     0,
		AppUnstakingBlocks:                       2016,
		AppMinimumPauseBlocks:                    4,
		AppMaxPauseBlocks:                        672,
		ServiceNodeMinimumStake:                  types.BigIntToString(big.NewInt(15000000000)),
		ServiceNodeMaxChains:                     15,
		ServiceNodeUnstakingBlocks:               2016,
		ServiceNodeMinimumPauseBlocks:            4,
		ServiceNodeMaxPauseBlocks:                672,
		ServiceNodesPerSession:                   24,
		FishermanMinimumStake:                    types.BigIntToString(big.NewInt(15000000000)),
		FishermanMaxChains:                       15,
		FishermanUnstakingBlocks:                 2016,
		FishermanMinimumPauseBlocks:              4,
		FishermanMaxPauseBlocks:                  672,
		ValidatorMinimumStake:                    types.BigIntToString(big.NewInt(15000000000)),
		ValidatorUnstakingBlocks:                 2016,
		ValidatorMinimumPauseBlocks:              4,
		ValidatorMaxPauseBlocks:                  672,
		ValidatorMaximumMissedBlocks:             5,
		ValidatorMaxEvidenceAgeInBlocks:          8,
		ProposerPercentageOfFees:                 10,
		MissedBlocksBurnPercentage:               1,
		DoubleSignBurnPercentage:                 5,
		MessageDoubleSignFee:                     types.BigIntToString(big.NewInt(10000)),
		MessageSendFee:                           types.BigIntToString(big.NewInt(10000)),
		MessageStakeFishermanFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeFishermanFee:             types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeFishermanFee:               types.BigIntToString(big.NewInt(10000)),
		MessagePauseFishermanFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseFishermanFee:               types.BigIntToString(big.NewInt(10000)),
		MessageFishermanPauseServiceNodeFee:      types.BigIntToString(big.NewInt(10000)),
		MessageTestScoreFee:                      types.BigIntToString(big.NewInt(10000)),
		MessageProveTestScoreFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageStakeAppFee:                       types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeAppFee:                   types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeAppFee:                     types.BigIntToString(big.NewInt(10000)),
		MessagePauseAppFee:                       types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseAppFee:                     types.BigIntToString(big.NewInt(10000)),
		MessageStakeValidatorFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeValidatorFee:             types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeValidatorFee:               types.BigIntToString(big.NewInt(10000)),
		MessagePauseValidatorFee:                 types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseValidatorFee:               types.BigIntToString(big.NewInt(10000)),
		MessageStakeServiceNodeFee:               types.BigIntToString(big.NewInt(10000)),
		MessageEditStakeServiceNodeFee:           types.BigIntToString(big.NewInt(10000)),
		MessageUnstakeServiceNodeFee:             types.BigIntToString(big.NewInt(10000)),
		MessagePauseServiceNodeFee:               types.BigIntToString(big.NewInt(10000)),
		MessageUnpauseServiceNodeFee:             types.BigIntToString(big.NewInt(10000)),
		MessageChangeParameterFee:                types.BigIntToString(big.NewInt(10000)),
		ACLOwner:                                 DefaultParamsOwner.Address(),
		BlocksPerSessionOwner:                    DefaultParamsOwner.Address(),
		AppMinimumStakeOwner:                     DefaultParamsOwner.Address(),
		AppMaxChainsOwner:                        DefaultParamsOwner.Address(),
		AppBaselineStakeRateOwner:                DefaultParamsOwner.Address(),
		AppStakingAdjustmentOwner:                DefaultParamsOwner.Address(),
		AppUnstakingBlocksOwner:                  DefaultParamsOwner.Address(),
		AppMinimumPauseBlocksOwner:               DefaultParamsOwner.Address(),
		AppMaxPausedBlocksOwner:                  DefaultParamsOwner.Address(),
		ServiceNodeMinimumStakeOwner:             DefaultParamsOwner.Address(),
		ServiceNodeMaxChainsOwner:                DefaultParamsOwner.Address(),
		ServiceNodeUnstakingBlocksOwner:          DefaultParamsOwner.Address(),
		ServiceNodeMinimumPauseBlocksOwner:       DefaultParamsOwner.Address(),
		ServiceNodeMaxPausedBlocksOwner:          DefaultParamsOwner.Address(),
		ServiceNodesPerSessionOwner:              DefaultParamsOwner.Address(),
		ParamFishermanMinimumStakeOwner:          DefaultParamsOwner.Address(),
		FishermanMaxChainsOwner:                  DefaultParamsOwner.Address(),
		FishermanUnstakingBlocksOwner:            DefaultParamsOwner.Address(),
		FishermanMinimumPauseBlocksOwner:         DefaultParamsOwner.Address(),
		FishermanMaxPausedBlocksOwner:            DefaultParamsOwner.Address(),
		ValidatorMinimumStakeOwner:               DefaultParamsOwner.Address(),
		ValidatorUnstakingBlocksOwner:            DefaultParamsOwner.Address(),
		ValidatorMinimumPauseBlocksOwner:         DefaultParamsOwner.Address(),
		ValidatorMaxPausedBlocksOwner:            DefaultParamsOwner.Address(),
		ValidatorMaximumMissedBlocksOwner:        DefaultParamsOwner.Address(),
		ValidatorMaxEvidenceAgeInBlocksOwner:     DefaultParamsOwner.Address(),
		ProposerPercentageOfFeesOwner:            DefaultParamsOwner.Address(),
		MissedBlocksBurnPercentageOwner:          DefaultParamsOwner.Address(),
		DoubleSignBurnPercentageOwner:            DefaultParamsOwner.Address(),
		MessageDoubleSignFeeOwner:                DefaultParamsOwner.Address(),
		MessageSendFeeOwner:                      DefaultParamsOwner.Address(),
		MessageStakeFishermanFeeOwner:            DefaultParamsOwner.Address(),
		MessageEditStakeFishermanFeeOwner:        DefaultParamsOwner.Address(),
		MessageUnstakeFishermanFeeOwner:          DefaultParamsOwner.Address(),
		MessagePauseFishermanFeeOwner:            DefaultParamsOwner.Address(),
		MessageUnpauseFishermanFeeOwner:          DefaultParamsOwner.Address(),
		MessageFishermanPauseServiceNodeFeeOwner: DefaultParamsOwner.Address(),
		MessageTestScoreFeeOwner:                 DefaultParamsOwner.Address(),
		MessageProveTestScoreFeeOwner:            DefaultParamsOwner.Address(),
		MessageStakeAppFeeOwner:                  DefaultParamsOwner.Address(),
		MessageEditStakeAppFeeOwner:              DefaultParamsOwner.Address(),
		MessageUnstakeAppFeeOwner:                DefaultParamsOwner.Address(),
		MessagePauseAppFeeOwner:                  DefaultParamsOwner.Address(),
		MessageUnpauseAppFeeOwner:                DefaultParamsOwner.Address(),
		MessageStakeValidatorFeeOwner:            DefaultParamsOwner.Address(),
		MessageEditStakeValidatorFeeOwner:        DefaultParamsOwner.Address(),
		MessageUnstakeValidatorFeeOwner:          DefaultParamsOwner.Address(),
		MessagePauseValidatorFeeOwner:            DefaultParamsOwner.Address(),
		MessageUnpauseValidatorFeeOwner:          DefaultParamsOwner.Address(),
		MessageStakeServiceNodeFeeOwner:          DefaultParamsOwner.Address(),
		MessageEditStakeServiceNodeFeeOwner:      DefaultParamsOwner.Address(),
		MessageUnstakeServiceNodeFeeOwner:        DefaultParamsOwner.Address(),
		MessagePauseServiceNodeFeeOwner:          DefaultParamsOwner.Address(),
		MessageUnpauseServiceNodeFeeOwner:        DefaultParamsOwner.Address(),
		MessageChangeParameterFeeOwner:           DefaultParamsOwner.Address(),
	}
}

func InsertParams(u *utility.UtilityContext, params *Params) types.Error {
	store := u.Store()
	if err := store.InitParams(); err != nil {
		return types.ErrInitParams(err)
	}
	err := store.SetBlocksPerSession(int(params.BlocksPerSession))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodesPerSession(int(params.ServiceNodesPerSession))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMaxAppChains(int(params.AppMaxChains))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetParamAppMinimumStake(params.AppMinimumStake)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetBaselineAppStakeRate(int(params.AppBaselineStakeRate))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetStakingAdjustment(int(params.AppStakingAdjustment))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppUnstakingBlocks(int(params.AppUnstakingBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppMinimumPauseBlocks(int(params.AppMinimumPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppMaxPausedBlocks(int(params.AppMaxPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetParamServiceNodeMinimumStake(params.ServiceNodeMinimumStake)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeMaxChains(int(params.ServiceNodeMaxChains))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeUnstakingBlocks(int(params.ServiceNodeUnstakingBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeMinimumPauseBlocks(int(params.ServiceNodeMinimumPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeMaxPausedBlocks(int(params.ServiceNodeMaxPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetParamFishermanMinimumStake(params.FishermanMinimumStake)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanMaxChains(int(params.FishermanMaxChains))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanUnstakingBlocks(int(params.FishermanUnstakingBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanMinimumPauseBlocks(int(params.FishermanMinimumPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanMaxPausedBlocks(int(params.FishermanMaxPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetParamValidatorMinimumStake(params.ValidatorMinimumStake)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorUnstakingBlocks(int(params.ValidatorUnstakingBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorMinimumPauseBlocks(int(params.ValidatorMinimumPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorMaxPausedBlocks(int(params.ValidatorMaxPauseBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorMaximumMissedBlocks(int(params.ValidatorMaximumMissedBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetProposerPercentageOfFees(int(params.ProposerPercentageOfFees))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMaxEvidenceAgeInBlocks(int(params.ValidatorMaxEvidenceAgeInBlocks))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMissedBlocksBurnPercentage(int(params.MissedBlocksBurnPercentage))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetDoubleSignBurnPercentage(int(params.DoubleSignBurnPercentage))
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetACLOwner(params.ACLOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetBlocksPerSessionOwner(params.BlocksPerSessionOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodesPerSessionOwner(params.ServiceNodesPerSessionOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMaxAppChainsOwner(params.AppMaxChainsOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppMinimumStakeOwner(params.AppMinimumStakeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetBaselineAppOwner(params.AppBaselineStakeRateOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetStakingAdjustmentOwner(params.AppStakingAdjustmentOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppUnstakingBlocksOwner(params.AppUnstakingBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppMinimumPauseBlocksOwner(params.AppMinimumPauseBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetAppMaxPausedBlocksOwner(params.AppMaxPausedBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetParamServiceNodeMinimumStakeOwner(params.ServiceNodeMinimumStakeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMaxServiceNodeChainsOwner(params.ServiceNodeMaxChainsOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeUnstakingBlocksOwner(params.ServiceNodeUnstakingBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeMinimumPauseBlocksOwner(params.ServiceNodeMinimumPauseBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetServiceNodeMaxPausedBlocksOwner(params.ServiceNodeMaxPausedBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanMinimumStakeOwner(params.ParamFishermanMinimumStakeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMaxFishermanChainsOwner(params.FishermanMaxChainsOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanUnstakingBlocksOwner(params.FishermanUnstakingBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanMinimumPauseBlocksOwner(params.ValidatorMinimumPauseBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetFishermanMaxPausedBlocksOwner(params.FishermanMaxPausedBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetParamValidatorMinimumStakeOwner(params.ValidatorMinimumStakeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorUnstakingBlocksOwner(params.FishermanUnstakingBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorMinimumPauseBlocksOwner(params.FishermanMinimumPauseBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorMaxPausedBlocksOwner(params.ValidatorMaxPausedBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetValidatorMaximumMissedBlocksOwner(params.ValidatorMaximumMissedBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetProposerPercentageOfFeesOwner(params.ProposerPercentageOfFeesOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMaxEvidenceAgeInBlocksOwner(params.ValidatorMaxEvidenceAgeInBlocksOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMissedBlocksBurnPercentageOwner(params.MissedBlocksBurnPercentageOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetDoubleSignBurnPercentageOwner(params.DoubleSignBurnPercentageOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageSendFeeOwner(params.MessageSendFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeFishermanFeeOwner(params.MessageStakeFishermanFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeFishermanFeeOwner(params.MessageEditStakeFishermanFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeFishermanFeeOwner(params.MessageUnstakeFishermanFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseFishermanFeeOwner(params.MessagePauseFishermanFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseFishermanFeeOwner(params.MessageUnpauseFishermanFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageFishermanPauseServiceNodeFeeOwner(params.MessageFishermanPauseServiceNodeFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageTestScoreFeeOwner(params.MessageTestScoreFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageProveTestScoreFeeOwner(params.MessageProveTestScoreFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeAppFeeOwner(params.MessageStakeAppFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeAppFeeOwner(params.MessageEditStakeAppFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeAppFeeOwner(params.MessageUnstakeAppFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseAppFeeOwner(params.MessagePauseAppFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseAppFeeOwner(params.MessageUnpauseAppFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeValidatorFeeOwner(params.MessageStakeValidatorFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeValidatorFeeOwner(params.MessageEditStakeValidatorFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeValidatorFeeOwner(params.MessageUnstakeValidatorFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseValidatorFeeOwner(params.MessagePauseValidatorFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseValidatorFeeOwner(params.MessageUnpauseValidatorFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeServiceNodeFeeOwner(params.MessageStakeServiceNodeFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeServiceNodeFeeOwner(params.MessageEditStakeServiceNodeFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeServiceNodeFeeOwner(params.MessageUnstakeServiceNodeFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseServiceNodeFeeOwner(params.MessagePauseServiceNodeFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseServiceNodeFeeOwner(params.MessageUnpauseServiceNodeFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageChangeParameterFeeOwner(params.MessageChangeParameterFeeOwner)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageSendFee(params.MessageSendFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeFishermanFee(params.MessageStakeFishermanFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeFishermanFee(params.MessageEditStakeFishermanFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeFishermanFee(params.MessageUnstakeFishermanFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseFishermanFee(params.MessagePauseFishermanFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseFishermanFee(params.MessageUnpauseFishermanFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageFishermanPauseServiceNodeFee(params.MessageFishermanPauseServiceNodeFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageTestScoreFee(params.MessageTestScoreFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageProveTestScoreFee(params.MessageProveTestScoreFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeAppFee(params.MessageStakeAppFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeAppFee(params.MessageEditStakeAppFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeAppFee(params.MessageUnstakeAppFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseAppFee(params.MessagePauseAppFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseAppFee(params.MessageUnpauseAppFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeValidatorFee(params.MessageStakeValidatorFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeValidatorFee(params.MessageEditStakeValidatorFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeValidatorFee(params.MessageUnstakeValidatorFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseValidatorFee(params.MessagePauseValidatorFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseValidatorFee(params.MessageUnpauseValidatorFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageStakeServiceNodeFee(params.MessageStakeServiceNodeFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageEditStakeServiceNodeFee(params.MessageEditStakeServiceNodeFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnstakeServiceNodeFee(params.MessageUnstakeServiceNodeFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessagePauseServiceNodeFee(params.MessagePauseServiceNodeFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageUnpauseServiceNodeFee(params.MessageUnpauseServiceNodeFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	err = store.SetMessageChangeParameterFee(params.MessageChangeParameterFee)
	if err != nil {
		return types.ErrUpdateParam(err)
	}
	return nil
}
