package persistence

import (
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(https://github.com/pokt-network/pocket/issues/76): Optimize gov parameters implementation & schema.

func (p PostgresContext) GetBlocksPerSession() (int, error) {
	return GetParam[int](p, types.BlocksPerSessionParamName)
}

func (p PostgresContext) GetParamAppMinimumStake() (string, error) {
	return GetParam[string](p, types.AppMinimumStakeParamName)
}

func (p PostgresContext) GetMaxAppChains() (int, error) {
	return GetParam[int](p, types.AppMaxChainsParamName)
}

func (p PostgresContext) GetBaselineAppStakeRate() (int, error) {
	return GetParam[int](p, types.AppBaselineStakeRateParamName)
}

func (p PostgresContext) GetStabilityAdjustment() (int, error) {
	return GetParam[int](p, types.AppStakingAdjustmentParamName)
}

func (p PostgresContext) GetAppUnstakingBlocks() (int, error) {
	return GetParam[int](p, types.AppUnstakingBlocksParamName)
}

func (p PostgresContext) GetAppMinimumPauseBlocks() (int, error) {
	return GetParam[int](p, types.AppMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetAppMaxPausedBlocks() (int, error) {
	return GetParam[int](p, types.AppMaxPauseBlocksParamName)
}

func (p PostgresContext) GetParamServiceNodeMinimumStake() (string, error) {
	return GetParam[string](p, types.ServiceNodeMinimumStakeParamName)
}

func (p PostgresContext) GetServiceNodeMaxChains() (int, error) {
	return GetParam[int](p, types.ServiceNodeMaxChainsParamName)
}

func (p PostgresContext) GetServiceNodeUnstakingBlocks() (int, error) {
	return GetParam[int](p, types.ServiceNodeUnstakingBlocksParamName)
}

func (p PostgresContext) GetServiceNodeMinimumPauseBlocks() (int, error) {
	return GetParam[int](p, types.ServiceNodeMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetServiceNodeMaxPausedBlocks() (int, error) {
	return GetParam[int](p, types.ServiceNodeMaxPauseBlocksParamName)
}

func (p PostgresContext) GetServiceNodesPerSession() (int, error) {
	return GetParam[int](p, types.ServiceNodesPerSessionParamName)
}

func (p PostgresContext) GetParamFishermanMinimumStake() (string, error) {
	return GetParam[string](p, types.FishermanMinimumStakeParamName)
}

func (p PostgresContext) GetFishermanMaxChains() (int, error) {
	return GetParam[int](p, types.FishermanMaxChainsParamName)
}

func (p PostgresContext) GetFishermanUnstakingBlocks() (int, error) {
	return GetParam[int](p, types.FishermanUnstakingBlocksParamName)
}

func (p PostgresContext) GetFishermanMinimumPauseBlocks() (int, error) {
	return GetParam[int](p, types.FishermanMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetFishermanMaxPausedBlocks() (int, error) {
	return GetParam[int](p, types.FishermanMaxPauseBlocksParamName)
}

func (p PostgresContext) GetParamValidatorMinimumStake() (string, error) {
	return GetParam[string](p, types.ValidatorMinimumStakeParamName)
}

func (p PostgresContext) GetValidatorUnstakingBlocks() (int, error) {
	return GetParam[int](p, types.ValidatorUnstakingBlocksParamName)
}

func (p PostgresContext) GetValidatorMinimumPauseBlocks() (int, error) {
	return GetParam[int](p, types.ValidatorMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetValidatorMaxPausedBlocks() (int, error) {
	return GetParam[int](p, types.ValidatorMaxPausedBlocksParamName)
}

func (p PostgresContext) GetValidatorMaximumMissedBlocks() (int, error) {
	return GetParam[int](p, types.ValidatorMaximumMissedBlocksParamName)
}

func (p PostgresContext) GetProposerPercentageOfFees() (int, error) {
	return GetParam[int](p, types.ProposerPercentageOfFeesParamName)
}

func (p PostgresContext) GetMaxEvidenceAgeInBlocks() (int, error) {
	return GetParam[int](p, types.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (p PostgresContext) GetMissedBlocksBurnPercentage() (int, error) {
	return GetParam[int](p, types.MissedBlocksBurnPercentageParamName)
}

func (p PostgresContext) GetDoubleSignBurnPercentage() (int, error) {
	return GetParam[int](p, types.DoubleSignBurnPercentageParamName)
}

func (p PostgresContext) GetMessageDoubleSignFee() (string, error) {
	return GetParam[string](p, types.MessageDoubleSignFee)
}

func (p PostgresContext) GetMessageSendFee() (string, error) {
	return GetParam[string](p, types.MessageSendFee)
}

func (p PostgresContext) GetMessageStakeFishermanFee() (string, error) {
	return GetParam[string](p, types.MessageStakeFishermanFee)
}

func (p PostgresContext) GetMessageEditStakeFishermanFee() (string, error) {
	return GetParam[string](p, types.MessageEditStakeFishermanFee)
}

func (p PostgresContext) GetMessageUnstakeFishermanFee() (string, error) {
	return GetParam[string](p, types.MessageUnstakeFishermanFee)
}

func (p PostgresContext) GetMessagePauseFishermanFee() (string, error) {
	return GetParam[string](p, types.MessagePauseFishermanFee)
}

func (p PostgresContext) GetMessageUnpauseFishermanFee() (string, error) {
	return GetParam[string](p, types.MessageUnpauseFishermanFee)
}

func (p PostgresContext) GetMessageFishermanPauseServiceNodeFee() (string, error) {
	return GetParam[string](p, types.MessagePauseServiceNodeFee)
}

func (p PostgresContext) GetMessageTestScoreFee() (string, error) {
	return GetParam[string](p, types.MessageTestScoreFee)
}

func (p PostgresContext) GetMessageProveTestScoreFee() (string, error) {
	return GetParam[string](p, types.MessageProveTestScoreFee)
}

func (p PostgresContext) GetMessageStakeAppFee() (string, error) {
	return GetParam[string](p, types.MessageEditStakeAppFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeAppFee() (string, error) {
	return GetParam[string](p, types.MessageEditStakeAppFee)
}

func (p PostgresContext) GetMessageUnstakeAppFee() (string, error) {
	return GetParam[string](p, types.MessageUnstakeAppFee)
}

func (p PostgresContext) GetMessagePauseAppFee() (string, error) {
	return GetParam[string](p, types.MessagePauseAppFee)
}

func (p PostgresContext) GetMessageUnpauseAppFee() (string, error) {
	return GetParam[string](p, types.MessageUnpauseAppFee)
}

func (p PostgresContext) GetMessageStakeValidatorFee() (string, error) {
	return GetParam[string](p, types.MessageStakeValidatorFee)
}

func (p PostgresContext) GetMessageEditStakeValidatorFee() (string, error) {
	return GetParam[string](p, types.MessageEditStakeValidatorFee)
}

func (p PostgresContext) GetMessageUnstakeValidatorFee() (string, error) {
	return GetParam[string](p, types.MessageUnstakeValidatorFee)
}

func (p PostgresContext) GetMessagePauseValidatorFee() (string, error) {
	return GetParam[string](p, types.MessagePauseValidatorFee)
}

func (p PostgresContext) GetMessageUnpauseValidatorFee() (string, error) {
	return GetParam[string](p, types.MessageUnpauseValidatorFee)
}

func (p PostgresContext) GetMessageStakeServiceNodeFee() (string, error) {
	return GetParam[string](p, types.MessageStakeServiceNodeFee)
}

func (p PostgresContext) GetMessageEditStakeServiceNodeFee() (string, error) {
	return GetParam[string](p, types.MessageEditStakeServiceNodeFee)
}

func (p PostgresContext) GetMessageUnstakeServiceNodeFee() (string, error) {
	return GetParam[string](p, types.MessageUnstakeServiceNodeFee)
}

func (p PostgresContext) GetMessagePauseServiceNodeFee() (string, error) {
	return GetParam[string](p, types.MessagePauseServiceNodeFee)
}

func (p PostgresContext) GetMessageUnpauseServiceNodeFee() (string, error) {
	return GetParam[string](p, types.MessageUnpauseServiceNodeFee)
}

func (p PostgresContext) GetMessageChangeParameterFee() (string, error) {
	return GetParam[string](p, types.MessageChangeParameterFee)
}

func (p PostgresContext) SetBlocksPerSession(i int) error {
	return SetParam(p, types.BlocksPerSessionParamName, i)
}

func (p PostgresContext) SetParamAppMinimumStake(i string) error {
	return SetParam(p, types.AppMinimumStakeParamName, i)
}

func (p PostgresContext) SetMaxAppChains(i int) error {
	return SetParam(p, types.FishermanMaxChainsParamName, i)
}

func (p PostgresContext) SetBaselineAppStakeRate(i int) error {
	return SetParam(p, types.AppBaselineStakeRateParamName, i)
}

func (p PostgresContext) SetStakingAdjustment(i int) error {
	return SetParam(p, types.AppStakingAdjustmentParamName, i)
}

func (p PostgresContext) SetAppUnstakingBlocks(i int) error {
	return SetParam(p, types.AppUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetAppMinimumPauseBlocks(i int) error {
	return SetParam(p, types.AppMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetAppMaxPausedBlocks(i int) error {
	return SetParam(p, types.AppMaxPauseBlocksParamName, i)
}

func (p PostgresContext) SetParamServiceNodeMinimumStake(i string) error {
	return SetParam(p, types.ServiceNodeMinimumStakeParamName, i)
}

func (p PostgresContext) SetServiceNodeMaxChains(i int) error {
	return SetParam(p, types.ServiceNodeMaxChainsParamName, i)
}

func (p PostgresContext) SetServiceNodeUnstakingBlocks(i int) error {
	return SetParam(p, types.ServiceNodeUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetServiceNodeMinimumPauseBlocks(i int) error {
	return SetParam(p, types.ServiceNodeMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetServiceNodeMaxPausedBlocks(i int) error {
	return SetParam(p, types.ServiceNodeMaxPauseBlocksParamName, i)
}

func (p PostgresContext) SetServiceNodesPerSession(i int) error {
	return SetParam(p, types.ServiceNodesPerSessionParamName, i)
}

func (p PostgresContext) SetParamFishermanMinimumStake(i string) error {
	return SetParam(p, types.FishermanMinimumStakeParamName, i)
}

func (p PostgresContext) SetFishermanMaxChains(i int) error {
	return SetParam(p, types.FishermanMaxChainsParamName, i)
}

func (p PostgresContext) SetFishermanUnstakingBlocks(i int) error {
	return SetParam(p, types.FishermanUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetFishermanMinimumPauseBlocks(i int) error {
	return SetParam(p, types.FishermanMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetFishermanMaxPausedBlocks(i int) error {
	return SetParam(p, types.FishermanMaxPauseBlocksParamName, i)
}

func (p PostgresContext) SetParamValidatorMinimumStake(i string) error {
	return SetParam(p, types.ValidatorMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetValidatorUnstakingBlocks(i int) error {
	return SetParam(p, types.ValidatorUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetValidatorMinimumPauseBlocks(i int) error {
	return SetParam(p, types.ValidatorMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetValidatorMaxPausedBlocks(i int) error {
	return SetParam(p, types.ValidatorMaxPausedBlocksParamName, i)
}

func (p PostgresContext) SetValidatorMaximumMissedBlocks(i int) error {
	return SetParam(p, types.ValidatorMaximumMissedBlocksParamName, i)
}

func (p PostgresContext) SetProposerPercentageOfFees(i int) error {
	return SetParam(p, types.ProposerPercentageOfFeesParamName, i)
}

func (p PostgresContext) SetMaxEvidenceAgeInBlocks(i int) error {
	return SetParam(p, types.ValidatorMaxEvidenceAgeInBlocksParamName, i)
}

func (p PostgresContext) SetMissedBlocksBurnPercentage(i int) error {
	return SetParam(p, types.MissedBlocksBurnPercentageParamName, i)
}

func (p PostgresContext) SetDoubleSignBurnPercentage(i int) error {
	return SetParam(p, types.DoubleSignBurnPercentageParamName, i)
}

func (p PostgresContext) SetMessageDoubleSignFee(i string) error {
	return SetParam(p, types.MessageDoubleSignFee, i)
}

func (p PostgresContext) SetMessageSendFee(i string) error {
	return SetParam(p, types.MessageSendFee, i)
}

func (p PostgresContext) SetMessageStakeFishermanFee(i string) error {
	return SetParam(p, types.MessageStakeFishermanFee, i)
}

func (p PostgresContext) SetMessageEditStakeFishermanFee(i string) error {
	return SetParam(p, types.MessageEditStakeFishermanFee, i)
}

func (p PostgresContext) SetMessageUnstakeFishermanFee(i string) error {
	return SetParam(p, types.MessageUnstakeFishermanFee, i)
}

func (p PostgresContext) SetMessagePauseFishermanFee(i string) error {
	return SetParam(p, types.MessagePauseFishermanFee, i)
}

func (p PostgresContext) SetMessageUnpauseFishermanFee(i string) error {
	return SetParam(p, types.MessageUnpauseFishermanFee, i)
}

func (p PostgresContext) SetMessageFishermanPauseServiceNodeFee(i string) error {
	return SetParam(p, types.MessagePauseServiceNodeFee, i)
}

func (p PostgresContext) SetMessageTestScoreFee(i string) error {
	return SetParam(p, types.MessageTestScoreFee, i)
}

func (p PostgresContext) SetMessageProveTestScoreFee(i string) error {
	return SetParam(p, types.MessageProveTestScoreFee, i)
}

func (p PostgresContext) SetMessageStakeAppFee(i string) error {
	return SetParam(p, types.MessageStakeAppFee, i)
}

func (p PostgresContext) SetMessageEditStakeAppFee(i string) error {
	return SetParam(p, types.MessageEditStakeAppFee, i)
}

func (p PostgresContext) SetMessageUnstakeAppFee(i string) error {
	return SetParam(p, types.MessageUnstakeAppFee, i)
}

func (p PostgresContext) SetMessagePauseAppFee(i string) error {
	return SetParam(p, types.MessagePauseAppFee, i)
}

func (p PostgresContext) SetMessageUnpauseAppFee(i string) error {
	return SetParam(p, types.MessageUnpauseAppFee, i)
}

func (p PostgresContext) SetMessageStakeValidatorFee(i string) error {
	return SetParam(p, types.MessageStakeValidatorFee, i)
}

func (p PostgresContext) SetMessageEditStakeValidatorFee(i string) error {
	return SetParam(p, types.MessageEditStakeValidatorFee, i)
}

func (p PostgresContext) SetMessageUnstakeValidatorFee(i string) error {
	return SetParam(p, types.MessageUnstakeValidatorFee, i)
}

func (p PostgresContext) SetMessagePauseValidatorFee(i string) error {
	return SetParam(p, types.MessagePauseValidatorFee, i)
}

func (p PostgresContext) SetMessageUnpauseValidatorFee(i string) error {
	return SetParam(p, types.MessageUnpauseValidatorFee, i)
}

func (p PostgresContext) SetMessageStakeServiceNodeFee(i string) error {
	return SetParam(p, types.MessageStakeServiceNodeFee, i)
}

func (p PostgresContext) SetMessageEditStakeServiceNodeFee(i string) error {
	return SetParam(p, types.MessageEditStakeServiceNodeFee, i)
}

func (p PostgresContext) SetMessageUnstakeServiceNodeFee(i string) error {
	return SetParam(p, types.MessageUnstakeServiceNodeFee, i)
}

func (p PostgresContext) SetMessagePauseServiceNodeFee(i string) error {
	return SetParam(p, types.MessagePauseServiceNodeFee, i)
}

func (p PostgresContext) SetMessageUnpauseServiceNodeFee(i string) error {
	return SetParam(p, types.MessageUnpauseServiceNodeFee, i)
}

func (p PostgresContext) SetMessageChangeParameterFee(i string) error {
	return SetParam(p, types.AppMinimumStakeParamName, i)
}

func (p PostgresContext) SetMessageDoubleSignFeeOwner(i []byte) error {
	return SetParam(p, types.MessageDoubleSignFeeOwner, i)
}

func (p PostgresContext) SetMessageSendFeeOwner(i []byte) error {
	return SetParam(p, types.MessageSendFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeFishermanFeeOwner(i []byte) error {
	return SetParam(p, types.MessageStakeFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeFishermanFeeOwner(i []byte) error {
	return SetParam(p, types.MessageEditStakeFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeFishermanFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnstakeFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseFishermanFeeOwner(i []byte) error {
	return SetParam(p, types.MessagePauseFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseFishermanFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnpauseFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageFishermanPauseServiceNodeFeeOwner(i []byte) error {
	return SetParam(p, types.MessageFishermanPauseServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageTestScoreFeeOwner(i []byte) error {
	return SetParam(p, types.MessageTestScoreFeeOwner, i)
}

func (p PostgresContext) SetMessageProveTestScoreFeeOwner(i []byte) error {
	return SetParam(p, types.MessageProveTestScoreFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeAppFeeOwner(i []byte) error {
	return SetParam(p, types.MessageStakeAppFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeAppFeeOwner(i []byte) error {
	return SetParam(p, types.MessageEditStakeAppFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeAppFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnstakeAppFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseAppFeeOwner(i []byte) error {
	return SetParam(p, types.MessagePauseAppFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseAppFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnpauseAppFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeValidatorFeeOwner(i []byte) error {
	return SetParam(p, types.MessageStakeValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeValidatorFeeOwner(i []byte) error {
	return SetParam(p, types.MessageEditStakeValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeValidatorFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnstakeValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseValidatorFeeOwner(i []byte) error {
	return SetParam(p, types.MessagePauseValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseValidatorFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnpauseValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeServiceNodeFeeOwner(i []byte) error {
	return SetParam(p, types.MessageStakeServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeServiceNodeFeeOwner(i []byte) error {
	return SetParam(p, types.MessageEditStakeServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeServiceNodeFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnstakeServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseServiceNodeFeeOwner(i []byte) error {
	return SetParam(p, types.MessagePauseServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseServiceNodeFeeOwner(i []byte) error {
	return SetParam(p, types.MessageUnpauseServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageChangeParameterFeeOwner(i []byte) error {
	return SetParam(p, types.MessageChangeParameterFeeOwner, i)
}

func (p PostgresContext) GetAclOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AclOwner)
}

func (p PostgresContext) SetAclOwner(i []byte) error {
	return SetParam(p, types.AclOwner, i)
}

func (p PostgresContext) SetBlocksPerSessionOwner(i []byte) error {
	return SetParam(p, types.BlocksPerSessionOwner, i)
}

func (p PostgresContext) GetBlocksPerSessionOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.BlocksPerSessionOwner)
}

func (p PostgresContext) GetMaxAppChainsOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppMaxChainsOwner)
}

func (p PostgresContext) SetMaxAppChainsOwner(i []byte) error {
	return SetParam(p, types.AppMaxChainsOwner, i)
}

func (p PostgresContext) GetAppMinimumStakeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppMinimumStakeOwner)
}

func (p PostgresContext) SetAppMinimumStakeOwner(i []byte) error {
	return SetParam(p, types.AppMinimumStakeOwner, i)
}

func (p PostgresContext) GetBaselineAppOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppBaselineStakeRateOwner)
}

func (p PostgresContext) SetBaselineAppOwner(i []byte) error {
	return SetParam(p, types.AppBaselineStakeRateOwner, i)
}

func (p PostgresContext) GetStakingAdjustmentOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppStakingAdjustmentOwner)
}

func (p PostgresContext) SetStakingAdjustmentOwner(i []byte) error {
	return SetParam(p, types.AppStakingAdjustmentOwner, i)
}

func (p PostgresContext) GetAppUnstakingBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppUnstakingBlocksOwner)
}

func (p PostgresContext) SetAppUnstakingBlocksOwner(i []byte) error {
	return SetParam(p, types.AppUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetAppMinimumPauseBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetAppMinimumPauseBlocksOwner(i []byte) error {
	return SetParam(p, types.AppMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetAppMaxPausedBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.AppMaxPausedBlocksOwner)
}

func (p PostgresContext) SetAppMaxPausedBlocksOwner(i []byte) error {
	return SetParam(p, types.AppMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetParamServiceNodeMinimumStakeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ServiceNodeMinimumStakeOwner)
}

func (p PostgresContext) SetServiceNodeMinimumStakeOwner(i []byte) error {
	return SetParam(p, types.ServiceNodeMinimumStakeOwner, i)
}

func (p PostgresContext) GetServiceNodeMaxChainsOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ServiceNodeMaxChainsOwner)
}

func (p PostgresContext) SetMaxServiceNodeChainsOwner(i []byte) error {
	return SetParam(p, types.ServiceNodeMaxChainsOwner, i)
}

func (p PostgresContext) GetServiceNodeUnstakingBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ServiceNodeUnstakingBlocksOwner)
}

func (p PostgresContext) SetServiceNodeUnstakingBlocksOwner(i []byte) error {
	return SetParam(p, types.ServiceNodeUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetServiceNodeMinimumPauseBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ServiceNodeMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetServiceNodeMinimumPauseBlocksOwner(i []byte) error {
	return SetParam(p, types.ServiceNodeMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetServiceNodeMaxPausedBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ServiceNodeMaxPausedBlocksOwner)
}

func (p PostgresContext) SetServiceNodeMaxPausedBlocksOwner(i []byte) error {
	return SetParam(p, types.ServiceNodeMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetFishermanMinimumStakeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.FishermanMinimumStakeOwner)
}

func (p PostgresContext) SetFishermanMinimumStakeOwner(i []byte) error {
	return SetParam(p, types.FishermanMinimumStakeOwner, i)
}

func (p PostgresContext) GetMaxFishermanChainsOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.FishermanMaxChainsOwner)
}

func (p PostgresContext) SetMaxFishermanChainsOwner(i []byte) error {
	return SetParam(p, types.FishermanMaxChainsOwner, i)
}

func (p PostgresContext) GetFishermanUnstakingBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.FishermanUnstakingBlocksOwner)
}

func (p PostgresContext) SetFishermanUnstakingBlocksOwner(i []byte) error {
	return SetParam(p, types.FishermanUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetFishermanMinimumPauseBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.FishermanMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetFishermanMinimumPauseBlocksOwner(i []byte) error {
	return SetParam(p, types.FishermanMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetFishermanMaxPausedBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.FishermanMaxPausedBlocksOwner)
}

func (p PostgresContext) SetFishermanMaxPausedBlocksOwner(i []byte) error {
	return SetParam(p, types.FishermanMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMinimumStakeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ValidatorMinimumStakeOwner)
}

func (p PostgresContext) SetValidatorMinimumStakeOwner(i []byte) error {
	return SetParam(p, types.ValidatorMinimumStakeOwner, i)
}

func (p PostgresContext) GetValidatorUnstakingBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ValidatorUnstakingBlocksOwner)
}

func (p PostgresContext) SetValidatorUnstakingBlocksOwner(i []byte) error {
	return SetParam(p, types.ValidatorUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMinimumPauseBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ValidatorMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetValidatorMinimumPauseBlocksOwner(i []byte) error {
	return SetParam(p, types.ValidatorMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMaxPausedBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ValidatorMaxPausedBlocksOwner)
}

func (p PostgresContext) SetValidatorMaxPausedBlocksOwner(i []byte) error {
	return SetParam(p, types.ValidatorMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMaximumMissedBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ValidatorMaximumMissedBlocksOwner)
}

func (p PostgresContext) SetValidatorMaximumMissedBlocksOwner(i []byte) error {
	return SetParam(p, types.ValidatorMaximumMissedBlocksOwner, i)
}

func (p PostgresContext) GetProposerPercentageOfFeesOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ProposerPercentageOfFeesOwner)
}

func (p PostgresContext) SetProposerPercentageOfFeesOwner(i []byte) error {
	return SetParam(p, types.ProposerPercentageOfFeesOwner, i)
}

func (p PostgresContext) GetMaxEvidenceAgeInBlocksOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ValidatorMaxEvidenceAgeInBlocksOwner)
}

func (p PostgresContext) SetMaxEvidenceAgeInBlocksOwner(i []byte) error {
	return SetParam(p, types.ValidatorMaxEvidenceAgeInBlocksOwner, i)
}

func (p PostgresContext) GetMissedBlocksBurnPercentageOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MissedBlocksBurnPercentageOwner)
}

func (p PostgresContext) SetMissedBlocksBurnPercentageOwner(i []byte) error {
	return SetParam(p, types.MissedBlocksBurnPercentageOwner, i)
}

func (p PostgresContext) GetDoubleSignBurnPercentageOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.DoubleSignBurnPercentageOwner)
}

func (p PostgresContext) SetDoubleSignBurnPercentageOwner(i []byte) error {
	return SetParam(p, types.DoubleSignBurnPercentageOwner, i)
}

func (p PostgresContext) SetServiceNodesPerSessionOwner(i []byte) error {
	return SetParam(p, types.ServiceNodesPerSessionOwner, i)
}

func (p PostgresContext) GetServiceNodesPerSessionOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.ServiceNodesPerSessionOwner)
}

func (p PostgresContext) GetMessageDoubleSignFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageDoubleSignFeeOwner)
}

func (p PostgresContext) GetMessageSendFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageSendFeeOwner)
}

func (p PostgresContext) GetMessageStakeFishermanFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageStakeFishermanFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeFishermanFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageEditStakeFishermanFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeFishermanFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnstakeFishermanFeeOwner)
}

func (p PostgresContext) GetMessagePauseFishermanFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessagePauseFishermanFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseFishermanFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnpauseFishermanFeeOwner)
}

func (p PostgresContext) GetMessageFishermanPauseServiceNodeFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageFishermanPauseServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageTestScoreFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageTestScoreFeeOwner)
}

func (p PostgresContext) GetMessageProveTestScoreFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageProveTestScoreFeeOwner)
}

func (p PostgresContext) GetMessageStakeAppFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageStakeAppFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeAppFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageEditStakeAppFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeAppFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnstakeAppFeeOwner)
}

func (p PostgresContext) GetMessagePauseAppFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessagePauseAppFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseAppFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnpauseAppFeeOwner)
}

func (p PostgresContext) GetMessageStakeValidatorFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageStakeValidatorFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeValidatorFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageEditStakeValidatorFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeValidatorFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnstakeValidatorFeeOwner)
}

func (p PostgresContext) GetMessagePauseValidatorFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessagePauseValidatorFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseValidatorFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnpauseValidatorFeeOwner)
}

func (p PostgresContext) GetMessageStakeServiceNodeFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageStakeServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeServiceNodeFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageEditStakeServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeServiceNodeFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnstakeServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessagePauseServiceNodeFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessagePauseServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseServiceNodeFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageUnpauseServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageChangeParameterFeeOwner() ([]byte, error) {
	return GetParam[[]byte](p, types.MessageChangeParameterFeeOwner)
}

func (p PostgresContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	return GetParam[int](p, types.ServiceNodesPerSessionParamName)
}

func (p PostgresContext) InitParams() error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertParams(genesis.DefaultParams()))
	return err
}

func SetParam[T schema.ParamTypes](p PostgresContext, paramName string, paramValue T) error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}
	height, err := p.GetHeight()
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.NullifyParamQuery(paramName, height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.SetParam(paramName, paramValue, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func GetParam[T int | string | []byte](p PostgresContext, paramName string) (i T, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return i, err
	}
	err = conn.QueryRow(ctx, schema.GetParamQuery(paramName)).Scan(&i)
	return
}
