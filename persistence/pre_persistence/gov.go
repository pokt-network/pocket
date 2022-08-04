package pre_persistence

import (
	"fmt"

	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"

	"github.com/pokt-network/pocket/shared/types"
)

func (m *PrePersistenceContext) InitParams() error {
	codec := types.GetCodec()
	db := m.Store()
	p := typesGenesis.DefaultParams()
	bz, err := codec.Marshal(p)
	if err != nil {
		return err
	}
	return db.Put(ParamsPrefixKey, bz)
}

func (m *PrePersistenceContext) GetParams(height int64) (p *typesGenesis.Params, err error) {
	p = &typesGenesis.Params{}
	codec := types.GetCodec()
	var paramsBz []byte
	if height == m.Height {
		db := m.Store()
		paramsBz, err = db.Get(ParamsPrefixKey)
		if err != nil {
			return nil, err
		}
	} else {
		paramsBz, err = m.Parent.GetCommitDB().Get(HeightKey(height, ParamsPrefixKey))
		if err != nil {
			return nil, nil
		}
	}
	if err := codec.Unmarshal(paramsBz, p); err != nil {
		return nil, err
	}
	return
}

func InsertPersistenceParams(store *PrePersistenceContext, params *typesGenesis.Params) types.Error {
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
	err = store.SetAclOwner(params.AclOwner)
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
	err = store.SetServiceNodeMinimumStakeOwner(params.ServiceNodeMinimumStakeOwner)
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
	err = store.SetFishermanMinimumStakeOwner(params.FishermanMinimumStakeOwner)
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
	err = store.SetValidatorMinimumStakeOwner(params.ValidatorMinimumStakeOwner)
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

// TODO: (@deblasis) added only to allow compilation since I changed the PersistenceContext interface.
// PrePersistence is about to be deprecated
func (m *PrePersistenceContext) GetBlocksPerSession(height int64) (int, error) {
	params, err := m.GetParams(height)
	if err != nil {
		return types.ZeroInt, err
	}
	return int(params.BlocksPerSession), nil
}

func (m *PrePersistenceContext) SetParams(p *typesGenesis.Params) error {
	codec := types.GetCodec()
	store := m.Store()
	bz, err := codec.Marshal(p)
	if err != nil {
		return err
	}
	return store.Put(ParamsPrefixKey, bz)
}

func (m *PrePersistenceContext) SetBlocksPerSession(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.BlocksPerSession = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetParamAppMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumStake = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMaxAppChains(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxChains = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetBaselineAppStakeRate(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppBaselineStakeRate = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetStakingAdjustment(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppStakingAdjustment = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetParamServiceNodeMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumStake = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeMaxChains(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxChains = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodesPerSession(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodesPerSession = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetParamFishermanMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumStake = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanMaxChains(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxChains = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetParamValidatorMinimumStake(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumStake = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorUnstakingBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorUnstakingBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMinimumPauseBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMaxPausedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxPauseBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMaximumMissedBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaximumMissedBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetProposerPercentageOfFees(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ProposerPercentageOfFees = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMaxEvidenceAgeInBlocks(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxEvidenceAgeInBlocks = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMissedBlocksBurnPercentage(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MissedBlocksBurnPercentage = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetDoubleSignBurnPercentage(i int) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.DoubleSignBurnPercentage = int32(i)
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageDoubleSignFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageDoubleSignFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageSendFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageSendFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeFishermanFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeFishermanFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeFishermanFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseFishermanFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseFishermanFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseFishermanFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageFishermanPauseServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseServiceNodeFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageTestScoreFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageTestScoreFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageProveTestScoreFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageProveTestScoreFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeAppFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeAppFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeAppFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseAppFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseAppFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseAppFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeValidatorFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeValidatorFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeValidatorFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseValidatorFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseValidatorFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseValidatorFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeServiceNodeFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeServiceNodeFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeServiceNodeFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageFishermanPauseServiceNodeFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseServiceNodeFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseServiceNodeFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageChangeParameterFee(s string) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageChangeParameterFee = s
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageDoubleSignFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageDoubleSignFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageSendFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageSendFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseFishermanFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseFishermanFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageFishermanPauseServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageFishermanPauseServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageTestScoreFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageTestScoreFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageProveTestScoreFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageProveTestScoreFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseAppFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseAppFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseValidatorFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseValidatorFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageStakeServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageStakeServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageEditStakeServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageEditStakeServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnstakeServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnstakeServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessagePauseServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessagePauseServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageUnpauseServiceNodeFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageUnpauseServiceNodeFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMessageChangeParameterFeeOwner(bytes []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MessageChangeParameterFeeOwner = bytes
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAclOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AclOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetBlocksPerSessionOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.BlocksPerSessionOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMaxAppChainsOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxChainsOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetBaselineAppOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppBaselineStakeRateOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetStakingAdjustmentOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppStakingAdjustmentOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMinimumPauseBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetAppMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.AppMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMaxServiceNodeChainsOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxChainsOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodeMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodeMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMaxFishermanChainsOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxChainsOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMinimumPauseBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetFishermanMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.FishermanMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMinimumStakeOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumStakeOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorUnstakingBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorUnstakingBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMinimumPauseBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMinimumPauseBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMaxPausedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxPausedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetValidatorMaximumMissedBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaximumMissedBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetProposerPercentageOfFeesOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ProposerPercentageOfFeesOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMaxEvidenceAgeInBlocksOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ValidatorMaxEvidenceAgeInBlocksOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetMissedBlocksBurnPercentageOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.MissedBlocksBurnPercentageOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetDoubleSignBurnPercentageOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.DoubleSignBurnPercentageOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) SetServiceNodesPerSessionOwner(owner []byte) error {
	params, err := m.GetParams(m.Height)
	if err != nil {
		return err
	}
	params.ServiceNodesPerSessionOwner = owner
	return m.SetParams(params)
}

func (m *PrePersistenceContext) GetIntParam(paramName string, height int64) (int, error) {
	return 0, fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) GetStringParam(paramName string, height int64) (string, error) {
	return "", fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) GetBytesParam(paramName string, height int64) (param []byte, err error) {
	return nil, fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) SetParam(paramName string, value interface{}) error {
	return fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) InitFlags() error {
	//not implemented and it never will 😈 (PrePersistence is sunsetting)
	return fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) GetIntFlag(paramName string, height int64) (value int, enabled bool, err error) {
	//not implemented and it never will 😈 (PrePersistence is sunsetting)
	return value, enabled, fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) GetStringFlag(paramName string, height int64) (value string, enabled bool, err error) {
	//not implemented and it never will 😈 (PrePersistence is sunsetting)
	return value, enabled, fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) GetBytesFlag(paramName string, height int64) (value []byte, enabled bool, err error) {
	//not implemented and it never will 😈 (PrePersistence is sunsetting)
	return value, enabled, fmt.Errorf("Obsolete")
}

func (p *PrePersistenceContext) SetFlag(paramName string, value interface{}, enabled bool) error {
	//not implemented and it never will 😈 (PrePersistence is sunsetting)
	return fmt.Errorf("Obsolete")
}
