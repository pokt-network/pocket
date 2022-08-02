package persistence

import (
	"github.com/jackc/pgx/v4"
	"github.com/pokt-network/pocket/persistence/schema"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
)

// TODO(https://github.com/pokt-network/pocket/issues/76): Optimize gov parameters implementation & schema.

func (p PostgresContext) GetBlocksPerSession() (int, error) {
	return p.GetIntParam(types.BlocksPerSessionParamName)
}

func (p PostgresContext) GetParamAppMinimumStake() (string, error) {
	return p.GetStringParam(types.AppMinimumStakeParamName)
}

func (p PostgresContext) GetMaxAppChains() (int, error) {
	return p.GetIntParam(types.AppMaxChainsParamName)
}

func (p PostgresContext) GetBaselineAppStakeRate() (int, error) {
	return p.GetIntParam(types.AppBaselineStakeRateParamName)
}

func (p PostgresContext) GetStabilityAdjustment() (int, error) {
	return p.GetIntParam(types.AppStakingAdjustmentParamName)
}

func (p PostgresContext) GetAppUnstakingBlocks() (int, error) {
	return p.GetIntParam(types.AppUnstakingBlocksParamName)
}

func (p PostgresContext) GetAppMinimumPauseBlocks() (int, error) {
	return p.GetIntParam(types.AppMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetAppMaxPausedBlocks() (int, error) {
	return p.GetIntParam(types.AppMaxPauseBlocksParamName)
}

func (p PostgresContext) GetParamServiceNodeMinimumStake() (string, error) {
	return p.GetStringParam(types.ServiceNodeMinimumStakeParamName)
}

func (p PostgresContext) GetServiceNodeMaxChains() (int, error) {
	return p.GetIntParam(types.ServiceNodeMaxChainsParamName)
}

func (p PostgresContext) GetServiceNodeUnstakingBlocks() (int, error) {
	return p.GetIntParam(types.ServiceNodeUnstakingBlocksParamName)
}

func (p PostgresContext) GetServiceNodeMinimumPauseBlocks() (int, error) {
	return p.GetIntParam(types.ServiceNodeMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetServiceNodeMaxPausedBlocks() (int, error) {
	return p.GetIntParam(types.ServiceNodeMaxPauseBlocksParamName)
}

func (p PostgresContext) GetServiceNodesPerSession() (int, error) {
	return p.GetIntParam(types.ServiceNodesPerSessionParamName)
}

func (p PostgresContext) GetParamFishermanMinimumStake() (string, error) {
	return p.GetStringParam(types.FishermanMinimumStakeParamName)
}

func (p PostgresContext) GetFishermanMaxChains() (int, error) {
	return p.GetIntParam(types.FishermanMaxChainsParamName)
}

func (p PostgresContext) GetFishermanUnstakingBlocks() (int, error) {
	return p.GetIntParam(types.FishermanUnstakingBlocksParamName)
}

func (p PostgresContext) GetFishermanMinimumPauseBlocks() (int, error) {
	return p.GetIntParam(types.FishermanMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetFishermanMaxPausedBlocks() (int, error) {
	return p.GetIntParam(types.FishermanMaxPauseBlocksParamName)
}

func (p PostgresContext) GetParamValidatorMinimumStake() (string, error) {
	return p.GetStringParam(types.ValidatorMinimumStakeParamName)
}

func (p PostgresContext) GetValidatorUnstakingBlocks() (int, error) {
	return p.GetIntParam(types.ValidatorUnstakingBlocksParamName)
}

func (p PostgresContext) GetValidatorMinimumPauseBlocks() (int, error) {
	return p.GetIntParam(types.ValidatorMinimumPauseBlocksParamName)
}

func (p PostgresContext) GetValidatorMaxPausedBlocks() (int, error) {
	return p.GetIntParam(types.ValidatorMaxPausedBlocksParamName)
}

func (p PostgresContext) GetValidatorMaximumMissedBlocks() (int, error) {
	return p.GetIntParam(types.ValidatorMaximumMissedBlocksParamName)
}

func (p PostgresContext) GetProposerPercentageOfFees() (int, error) {
	return p.GetIntParam(types.ProposerPercentageOfFeesParamName)
}

func (p PostgresContext) GetMaxEvidenceAgeInBlocks() (int, error) {
	return p.GetIntParam(types.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (p PostgresContext) GetMissedBlocksBurnPercentage() (int, error) {
	return p.GetIntParam(types.MissedBlocksBurnPercentageParamName)
}

func (p PostgresContext) GetDoubleSignBurnPercentage() (int, error) {
	return p.GetIntParam(types.DoubleSignBurnPercentageParamName)
}

func (p PostgresContext) GetMessageDoubleSignFee() (string, error) {
	return p.GetStringParam(types.MessageDoubleSignFee)
}

func (p PostgresContext) GetMessageSendFee() (string, error) {
	return p.GetStringParam(types.MessageSendFee)
}

func (p PostgresContext) GetMessageStakeFishermanFee() (string, error) {
	return p.GetStringParam(types.MessageStakeFishermanFee)
}

func (p PostgresContext) GetMessageEditStakeFishermanFee() (string, error) {
	return p.GetStringParam(types.MessageEditStakeFishermanFee)
}

func (p PostgresContext) GetMessageUnstakeFishermanFee() (string, error) {
	return p.GetStringParam(types.MessageUnstakeFishermanFee)
}

func (p PostgresContext) GetMessagePauseFishermanFee() (string, error) {
	return p.GetStringParam(types.MessagePauseFishermanFee)
}

func (p PostgresContext) GetMessageUnpauseFishermanFee() (string, error) {
	return p.GetStringParam(types.MessageUnpauseFishermanFee)
}

func (p PostgresContext) GetMessageFishermanPauseServiceNodeFee() (string, error) {
	return p.GetStringParam(types.MessagePauseServiceNodeFee)
}

func (p PostgresContext) GetMessageTestScoreFee() (string, error) {
	return p.GetStringParam(types.MessageTestScoreFee)
}

func (p PostgresContext) GetMessageProveTestScoreFee() (string, error) {
	return p.GetStringParam(types.MessageProveTestScoreFee)
}

func (p PostgresContext) GetMessageStakeAppFee() (string, error) {
	return p.GetStringParam(types.MessageEditStakeAppFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeAppFee() (string, error) {
	return p.GetStringParam(types.MessageEditStakeAppFee)
}

func (p PostgresContext) GetMessageUnstakeAppFee() (string, error) {
	return p.GetStringParam(types.MessageUnstakeAppFee)
}

func (p PostgresContext) GetMessagePauseAppFee() (string, error) {
	return p.GetStringParam(types.MessagePauseAppFee)
}

func (p PostgresContext) GetMessageUnpauseAppFee() (string, error) {
	return p.GetStringParam(types.MessageUnpauseAppFee)
}

func (p PostgresContext) GetMessageStakeValidatorFee() (string, error) {
	return p.GetStringParam(types.MessageStakeValidatorFee)
}

func (p PostgresContext) GetMessageEditStakeValidatorFee() (string, error) {
	return p.GetStringParam(types.MessageEditStakeValidatorFee)
}

func (p PostgresContext) GetMessageUnstakeValidatorFee() (string, error) {
	return p.GetStringParam(types.MessageUnstakeValidatorFee)
}

func (p PostgresContext) GetMessagePauseValidatorFee() (string, error) {
	return p.GetStringParam(types.MessagePauseValidatorFee)
}

func (p PostgresContext) GetMessageUnpauseValidatorFee() (string, error) {
	return p.GetStringParam(types.MessageUnpauseValidatorFee)
}

func (p PostgresContext) GetMessageStakeServiceNodeFee() (string, error) {
	return p.GetStringParam(types.MessageStakeServiceNodeFee)
}

func (p PostgresContext) GetMessageEditStakeServiceNodeFee() (string, error) {
	return p.GetStringParam(types.MessageEditStakeServiceNodeFee)
}

func (p PostgresContext) GetMessageUnstakeServiceNodeFee() (string, error) {
	return p.GetStringParam(types.MessageUnstakeServiceNodeFee)
}

func (p PostgresContext) GetMessagePauseServiceNodeFee() (string, error) {
	return p.GetStringParam(types.MessagePauseServiceNodeFee)
}

func (p PostgresContext) GetMessageUnpauseServiceNodeFee() (string, error) {
	return p.GetStringParam(types.MessageUnpauseServiceNodeFee)
}

func (p PostgresContext) GetMessageChangeParameterFee() (string, error) {
	return p.GetStringParam(types.MessageChangeParameterFee)
}

func (p PostgresContext) SetBlocksPerSession(i int) error {
	return p.SetParam(types.BlocksPerSessionParamName, i)
}

func (p PostgresContext) SetParamAppMinimumStake(i string) error {
	return p.SetParam(types.AppMinimumStakeParamName, i)
}

func (p PostgresContext) SetMaxAppChains(i int) error {
	return p.SetParam(types.FishermanMaxChainsParamName, i)
}

func (p PostgresContext) SetBaselineAppStakeRate(i int) error {
	return p.SetParam(types.AppBaselineStakeRateParamName, i)
}

func (p PostgresContext) SetStakingAdjustment(i int) error {
	return p.SetParam(types.AppStakingAdjustmentParamName, i)
}

func (p PostgresContext) SetAppUnstakingBlocks(i int) error {
	return p.SetParam(types.AppUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetAppMinimumPauseBlocks(i int) error {
	return p.SetParam(types.AppMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetAppMaxPausedBlocks(i int) error {
	return p.SetParam(types.AppMaxPauseBlocksParamName, i)
}

func (p PostgresContext) SetParamServiceNodeMinimumStake(i string) error {
	return p.SetParam(types.ServiceNodeMinimumStakeParamName, i)
}

func (p PostgresContext) SetServiceNodeMaxChains(i int) error {
	return p.SetParam(types.ServiceNodeMaxChainsParamName, i)
}

func (p PostgresContext) SetServiceNodeUnstakingBlocks(i int) error {
	return p.SetParam(types.ServiceNodeUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetServiceNodeMinimumPauseBlocks(i int) error {
	return p.SetParam(types.ServiceNodeMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetServiceNodeMaxPausedBlocks(i int) error {
	return p.SetParam(types.ServiceNodeMaxPauseBlocksParamName, i)
}

func (p PostgresContext) SetServiceNodesPerSession(i int) error {
	return p.SetParam(types.ServiceNodesPerSessionParamName, i)
}

func (p PostgresContext) SetParamFishermanMinimumStake(i string) error {
	return p.SetParam(types.FishermanMinimumStakeParamName, i)
}

func (p PostgresContext) SetFishermanMaxChains(i int) error {
	return p.SetParam(types.FishermanMaxChainsParamName, i)
}

func (p PostgresContext) SetFishermanUnstakingBlocks(i int) error {
	return p.SetParam(types.FishermanUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetFishermanMinimumPauseBlocks(i int) error {
	return p.SetParam(types.FishermanMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetFishermanMaxPausedBlocks(i int) error {
	return p.SetParam(types.FishermanMaxPauseBlocksParamName, i)
}

func (p PostgresContext) SetParamValidatorMinimumStake(i string) error {
	return p.SetParam(types.ValidatorMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetValidatorUnstakingBlocks(i int) error {
	return p.SetParam(types.ValidatorUnstakingBlocksParamName, i)
}

func (p PostgresContext) SetValidatorMinimumPauseBlocks(i int) error {
	return p.SetParam(types.ValidatorMinimumPauseBlocksParamName, i)
}

func (p PostgresContext) SetValidatorMaxPausedBlocks(i int) error {
	return p.SetParam(types.ValidatorMaxPausedBlocksParamName, i)
}

func (p PostgresContext) SetValidatorMaximumMissedBlocks(i int) error {
	return p.SetParam(types.ValidatorMaximumMissedBlocksParamName, i)
}

func (p PostgresContext) SetProposerPercentageOfFees(i int) error {
	return p.SetParam(types.ProposerPercentageOfFeesParamName, i)
}

func (p PostgresContext) SetMaxEvidenceAgeInBlocks(i int) error {
	return p.SetParam(types.ValidatorMaxEvidenceAgeInBlocksParamName, i)
}

func (p PostgresContext) SetMissedBlocksBurnPercentage(i int) error {
	return p.SetParam(types.MissedBlocksBurnPercentageParamName, i)
}

func (p PostgresContext) SetDoubleSignBurnPercentage(i int) error {
	return p.SetParam(types.DoubleSignBurnPercentageParamName, i)
}

func (p PostgresContext) SetMessageDoubleSignFee(i string) error {
	return p.SetParam(types.MessageDoubleSignFee, i)
}

func (p PostgresContext) SetMessageSendFee(i string) error {
	return p.SetParam(types.MessageSendFee, i)
}

func (p PostgresContext) SetMessageStakeFishermanFee(i string) error {
	return p.SetParam(types.MessageStakeFishermanFee, i)
}

func (p PostgresContext) SetMessageEditStakeFishermanFee(i string) error {
	return p.SetParam(types.MessageEditStakeFishermanFee, i)
}

func (p PostgresContext) SetMessageUnstakeFishermanFee(i string) error {
	return p.SetParam(types.MessageUnstakeFishermanFee, i)
}

func (p PostgresContext) SetMessagePauseFishermanFee(i string) error {
	return p.SetParam(types.MessagePauseFishermanFee, i)
}

func (p PostgresContext) SetMessageUnpauseFishermanFee(i string) error {
	return p.SetParam(types.MessageUnpauseFishermanFee, i)
}

func (p PostgresContext) SetMessageFishermanPauseServiceNodeFee(i string) error {
	return p.SetParam(types.MessagePauseServiceNodeFee, i)
}

func (p PostgresContext) SetMessageTestScoreFee(i string) error {
	return p.SetParam(types.MessageTestScoreFee, i)
}

func (p PostgresContext) SetMessageProveTestScoreFee(i string) error {
	return p.SetParam(types.MessageProveTestScoreFee, i)
}

func (p PostgresContext) SetMessageStakeAppFee(i string) error {
	return p.SetParam(types.MessageStakeAppFee, i)
}

func (p PostgresContext) SetMessageEditStakeAppFee(i string) error {
	return p.SetParam(types.MessageEditStakeAppFee, i)
}

func (p PostgresContext) SetMessageUnstakeAppFee(i string) error {
	return p.SetParam(types.MessageUnstakeAppFee, i)
}

func (p PostgresContext) SetMessagePauseAppFee(i string) error {
	return p.SetParam(types.MessagePauseAppFee, i)
}

func (p PostgresContext) SetMessageUnpauseAppFee(i string) error {
	return p.SetParam(types.MessageUnpauseAppFee, i)
}

func (p PostgresContext) SetMessageStakeValidatorFee(i string) error {
	return p.SetParam(types.MessageStakeValidatorFee, i)
}

func (p PostgresContext) SetMessageEditStakeValidatorFee(i string) error {
	return p.SetParam(types.MessageEditStakeValidatorFee, i)
}

func (p PostgresContext) SetMessageUnstakeValidatorFee(i string) error {
	return p.SetParam(types.MessageUnstakeValidatorFee, i)
}

func (p PostgresContext) SetMessagePauseValidatorFee(i string) error {
	return p.SetParam(types.MessagePauseValidatorFee, i)
}

func (p PostgresContext) SetMessageUnpauseValidatorFee(i string) error {
	return p.SetParam(types.MessageUnpauseValidatorFee, i)
}

func (p PostgresContext) SetMessageStakeServiceNodeFee(i string) error {
	return p.SetParam(types.MessageStakeServiceNodeFee, i)
}

func (p PostgresContext) SetMessageEditStakeServiceNodeFee(i string) error {
	return p.SetParam(types.MessageEditStakeServiceNodeFee, i)
}

func (p PostgresContext) SetMessageUnstakeServiceNodeFee(i string) error {
	return p.SetParam(types.MessageUnstakeServiceNodeFee, i)
}

func (p PostgresContext) SetMessagePauseServiceNodeFee(i string) error {
	return p.SetParam(types.MessagePauseServiceNodeFee, i)
}

func (p PostgresContext) SetMessageUnpauseServiceNodeFee(i string) error {
	return p.SetParam(types.MessageUnpauseServiceNodeFee, i)
}

func (p PostgresContext) SetMessageChangeParameterFee(i string) error {
	return p.SetParam(types.AppMinimumStakeParamName, i)
}

func (p PostgresContext) SetMessageDoubleSignFeeOwner(i []byte) error {
	return p.SetParam(types.MessageDoubleSignFeeOwner, i)
}

func (p PostgresContext) SetMessageSendFeeOwner(i []byte) error {
	return p.SetParam(types.MessageSendFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeFishermanFeeOwner(i []byte) error {
	return p.SetParam(types.MessageStakeFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeFishermanFeeOwner(i []byte) error {
	return p.SetParam(types.MessageEditStakeFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeFishermanFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnstakeFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseFishermanFeeOwner(i []byte) error {
	return p.SetParam(types.MessagePauseFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseFishermanFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnpauseFishermanFeeOwner, i)
}

func (p PostgresContext) SetMessageFishermanPauseServiceNodeFeeOwner(i []byte) error {
	return p.SetParam(types.MessageFishermanPauseServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageTestScoreFeeOwner(i []byte) error {
	return p.SetParam(types.MessageTestScoreFeeOwner, i)
}

func (p PostgresContext) SetMessageProveTestScoreFeeOwner(i []byte) error {
	return p.SetParam(types.MessageProveTestScoreFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeAppFeeOwner(i []byte) error {
	return p.SetParam(types.MessageStakeAppFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeAppFeeOwner(i []byte) error {
	return p.SetParam(types.MessageEditStakeAppFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeAppFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnstakeAppFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseAppFeeOwner(i []byte) error {
	return p.SetParam(types.MessagePauseAppFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseAppFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnpauseAppFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeValidatorFeeOwner(i []byte) error {
	return p.SetParam(types.MessageStakeValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeValidatorFeeOwner(i []byte) error {
	return p.SetParam(types.MessageEditStakeValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeValidatorFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnstakeValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseValidatorFeeOwner(i []byte) error {
	return p.SetParam(types.MessagePauseValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseValidatorFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnpauseValidatorFeeOwner, i)
}

func (p PostgresContext) SetMessageStakeServiceNodeFeeOwner(i []byte) error {
	return p.SetParam(types.MessageStakeServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageEditStakeServiceNodeFeeOwner(i []byte) error {
	return p.SetParam(types.MessageEditStakeServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageUnstakeServiceNodeFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnstakeServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessagePauseServiceNodeFeeOwner(i []byte) error {
	return p.SetParam(types.MessagePauseServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageUnpauseServiceNodeFeeOwner(i []byte) error {
	return p.SetParam(types.MessageUnpauseServiceNodeFeeOwner, i)
}

func (p PostgresContext) SetMessageChangeParameterFeeOwner(i []byte) error {
	return p.SetParam(types.MessageChangeParameterFeeOwner, i)
}

func (p PostgresContext) GetAclOwner() ([]byte, error) {
	return p.GetBytesParam(types.AclOwner)
}

func (p PostgresContext) SetAclOwner(i []byte) error {
	return p.SetParam(types.AclOwner, i)
}

func (p PostgresContext) SetBlocksPerSessionOwner(i []byte) error {
	return p.SetParam(types.BlocksPerSessionOwner, i)
}

func (p PostgresContext) GetBlocksPerSessionOwner() ([]byte, error) {
	return p.GetBytesParam(types.BlocksPerSessionOwner)
}

func (p PostgresContext) GetMaxAppChainsOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppMaxChainsOwner)
}

func (p PostgresContext) SetMaxAppChainsOwner(i []byte) error {
	return p.SetParam(types.AppMaxChainsOwner, i)
}

func (p PostgresContext) GetAppMinimumStakeOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppMinimumStakeOwner)
}

func (p PostgresContext) SetAppMinimumStakeOwner(i []byte) error {
	return p.SetParam(types.AppMinimumStakeOwner, i)
}

func (p PostgresContext) GetBaselineAppOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppBaselineStakeRateOwner)
}

func (p PostgresContext) SetBaselineAppOwner(i []byte) error {
	return p.SetParam(types.AppBaselineStakeRateOwner, i)
}

func (p PostgresContext) GetStakingAdjustmentOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppStakingAdjustmentOwner)
}

func (p PostgresContext) SetStakingAdjustmentOwner(i []byte) error {
	return p.SetParam(types.AppStakingAdjustmentOwner, i)
}

func (p PostgresContext) GetAppUnstakingBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppUnstakingBlocksOwner)
}

func (p PostgresContext) SetAppUnstakingBlocksOwner(i []byte) error {
	return p.SetParam(types.AppUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetAppMinimumPauseBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetAppMinimumPauseBlocksOwner(i []byte) error {
	return p.SetParam(types.AppMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetAppMaxPausedBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.AppMaxPausedBlocksOwner)
}

func (p PostgresContext) SetAppMaxPausedBlocksOwner(i []byte) error {
	return p.SetParam(types.AppMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetParamServiceNodeMinimumStakeOwner() ([]byte, error) {
	return p.GetBytesParam(types.ServiceNodeMinimumStakeOwner)
}

func (p PostgresContext) SetServiceNodeMinimumStakeOwner(i []byte) error {
	return p.SetParam(types.ServiceNodeMinimumStakeOwner, i)
}

func (p PostgresContext) GetServiceNodeMaxChainsOwner() ([]byte, error) {
	return p.GetBytesParam(types.ServiceNodeMaxChainsOwner)
}

func (p PostgresContext) SetMaxServiceNodeChainsOwner(i []byte) error {
	return p.SetParam(types.ServiceNodeMaxChainsOwner, i)
}

func (p PostgresContext) GetServiceNodeUnstakingBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ServiceNodeUnstakingBlocksOwner)
}

func (p PostgresContext) SetServiceNodeUnstakingBlocksOwner(i []byte) error {
	return p.SetParam(types.ServiceNodeUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetServiceNodeMinimumPauseBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ServiceNodeMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetServiceNodeMinimumPauseBlocksOwner(i []byte) error {
	return p.SetParam(types.ServiceNodeMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetServiceNodeMaxPausedBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ServiceNodeMaxPausedBlocksOwner)
}

func (p PostgresContext) SetServiceNodeMaxPausedBlocksOwner(i []byte) error {
	return p.SetParam(types.ServiceNodeMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetFishermanMinimumStakeOwner() ([]byte, error) {
	return p.GetBytesParam(types.FishermanMinimumStakeOwner)
}

func (p PostgresContext) SetFishermanMinimumStakeOwner(i []byte) error {
	return p.SetParam(types.FishermanMinimumStakeOwner, i)
}

func (p PostgresContext) GetMaxFishermanChainsOwner() ([]byte, error) {
	return p.GetBytesParam(types.FishermanMaxChainsOwner)
}

func (p PostgresContext) SetMaxFishermanChainsOwner(i []byte) error {
	return p.SetParam(types.FishermanMaxChainsOwner, i)
}

func (p PostgresContext) GetFishermanUnstakingBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.FishermanUnstakingBlocksOwner)
}

func (p PostgresContext) SetFishermanUnstakingBlocksOwner(i []byte) error {
	return p.SetParam(types.FishermanUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetFishermanMinimumPauseBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.FishermanMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetFishermanMinimumPauseBlocksOwner(i []byte) error {
	return p.SetParam(types.FishermanMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetFishermanMaxPausedBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.FishermanMaxPausedBlocksOwner)
}

func (p PostgresContext) SetFishermanMaxPausedBlocksOwner(i []byte) error {
	return p.SetParam(types.FishermanMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMinimumStakeOwner() ([]byte, error) {
	return p.GetBytesParam(types.ValidatorMinimumStakeOwner)
}

func (p PostgresContext) SetValidatorMinimumStakeOwner(i []byte) error {
	return p.SetParam(types.ValidatorMinimumStakeOwner, i)
}

func (p PostgresContext) GetValidatorUnstakingBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ValidatorUnstakingBlocksOwner)
}

func (p PostgresContext) SetValidatorUnstakingBlocksOwner(i []byte) error {
	return p.SetParam(types.ValidatorUnstakingBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMinimumPauseBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ValidatorMinimumPauseBlocksOwner)
}

func (p PostgresContext) SetValidatorMinimumPauseBlocksOwner(i []byte) error {
	return p.SetParam(types.ValidatorMinimumPauseBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMaxPausedBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ValidatorMaxPausedBlocksOwner)
}

func (p PostgresContext) SetValidatorMaxPausedBlocksOwner(i []byte) error {
	return p.SetParam(types.ValidatorMaxPausedBlocksOwner, i)
}

func (p PostgresContext) GetValidatorMaximumMissedBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ValidatorMaximumMissedBlocksOwner)
}

func (p PostgresContext) SetValidatorMaximumMissedBlocksOwner(i []byte) error {
	return p.SetParam(types.ValidatorMaximumMissedBlocksOwner, i)
}

func (p PostgresContext) GetProposerPercentageOfFeesOwner() ([]byte, error) {
	return p.GetBytesParam(types.ProposerPercentageOfFeesOwner)
}

func (p PostgresContext) SetProposerPercentageOfFeesOwner(i []byte) error {
	return p.SetParam(types.ProposerPercentageOfFeesOwner, i)
}

func (p PostgresContext) GetMaxEvidenceAgeInBlocksOwner() ([]byte, error) {
	return p.GetBytesParam(types.ValidatorMaxEvidenceAgeInBlocksOwner)
}

func (p PostgresContext) SetMaxEvidenceAgeInBlocksOwner(i []byte) error {
	return p.SetParam(types.ValidatorMaxEvidenceAgeInBlocksOwner, i)
}

func (p PostgresContext) GetMissedBlocksBurnPercentageOwner() ([]byte, error) {
	return p.GetBytesParam(types.MissedBlocksBurnPercentageOwner)
}

func (p PostgresContext) SetMissedBlocksBurnPercentageOwner(i []byte) error {
	return p.SetParam(types.MissedBlocksBurnPercentageOwner, i)
}

func (p PostgresContext) GetDoubleSignBurnPercentageOwner() ([]byte, error) {
	return p.GetBytesParam(types.DoubleSignBurnPercentageOwner)
}

func (p PostgresContext) SetDoubleSignBurnPercentageOwner(i []byte) error {
	return p.SetParam(types.DoubleSignBurnPercentageOwner, i)
}

func (p PostgresContext) SetServiceNodesPerSessionOwner(i []byte) error {
	return p.SetParam(types.ServiceNodesPerSessionOwner, i)
}

func (p PostgresContext) GetServiceNodesPerSessionOwner() ([]byte, error) {
	return p.GetBytesParam(types.ServiceNodesPerSessionOwner)
}

func (p PostgresContext) GetMessageDoubleSignFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageDoubleSignFeeOwner)
}

func (p PostgresContext) GetMessageSendFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageSendFeeOwner)
}

func (p PostgresContext) GetMessageStakeFishermanFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageStakeFishermanFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeFishermanFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageEditStakeFishermanFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeFishermanFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnstakeFishermanFeeOwner)
}

func (p PostgresContext) GetMessagePauseFishermanFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessagePauseFishermanFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseFishermanFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnpauseFishermanFeeOwner)
}

func (p PostgresContext) GetMessageFishermanPauseServiceNodeFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageFishermanPauseServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageTestScoreFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageTestScoreFeeOwner)
}

func (p PostgresContext) GetMessageProveTestScoreFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageProveTestScoreFeeOwner)
}

func (p PostgresContext) GetMessageStakeAppFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageStakeAppFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeAppFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageEditStakeAppFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeAppFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnstakeAppFeeOwner)
}

func (p PostgresContext) GetMessagePauseAppFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessagePauseAppFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseAppFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnpauseAppFeeOwner)
}

func (p PostgresContext) GetMessageStakeValidatorFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageStakeValidatorFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeValidatorFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageEditStakeValidatorFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeValidatorFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnstakeValidatorFeeOwner)
}

func (p PostgresContext) GetMessagePauseValidatorFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessagePauseValidatorFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseValidatorFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnpauseValidatorFeeOwner)
}

func (p PostgresContext) GetMessageStakeServiceNodeFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageStakeServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageEditStakeServiceNodeFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageEditStakeServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageUnstakeServiceNodeFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnstakeServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessagePauseServiceNodeFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessagePauseServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageUnpauseServiceNodeFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageUnpauseServiceNodeFeeOwner)
}

func (p PostgresContext) GetMessageChangeParameterFeeOwner() ([]byte, error) {
	return p.GetBytesParam(types.MessageChangeParameterFeeOwner)
}

func (p PostgresContext) GetServiceNodesPerSessionAt(height int64) (int, error) {
	return p.GetIntParam(types.ServiceNodesPerSessionParamName)
}

func (p PostgresContext) InitParams() error {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, schema.InsertParams(genesis.DefaultParams()))
	return err
}

// IMPROVE(team): Switch to generics
func (p PostgresContext) SetParam(paramName string, paramValue interface{}) error {
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
	if _, err = tx.Exec(ctx, schema.NullifyParamsQuery(height)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, schema.SetParam(paramName, paramValue, height)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p PostgresContext) GetIntParam(paramName string) (i int, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return 0, err
	}
	err = conn.QueryRow(ctx, schema.GetParamQuery(paramName)).Scan(&i)
	return
}

func (p PostgresContext) GetStringParam(paramName string) (s string, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return "", err
	}
	err = conn.QueryRow(ctx, schema.GetParamQuery(paramName)).Scan(&s)
	return
}

func (p PostgresContext) GetBytesParam(paramName string) (param []byte, err error) {
	ctx, conn, err := p.GetCtxAndConnection()
	if err != nil {
		return nil, err
	}
	err = conn.QueryRow(ctx, schema.GetParamQuery(paramName)).Scan(&param)
	return
}
