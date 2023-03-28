package unit_of_work

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TODO : After we change the interface to pass param name, simply use reflection to
//  iterate over all the params and test them. Suggestion: [Google's go-cmp] (https://github.com/google/go-cmp)

func DefaultTestingParams(_ *testing.T) *genesis.Params {
	return test_artifacts.DefaultParams()
}

func TestUtilityUnitOfWork_GetAppMaxChains(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	maxChains, err := uow.getAppMaxChains()
	require.NoError(t, err)
	require.Equal(t, int(defaultParams.GetAppMaxChains()), maxChains)
}

func TestUtilityUnitOfWork_GetAppMaxPausedBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	gotParam, err := uow.getAppMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, int(defaultParams.GetAppMaxPauseBlocks()), gotParam)
}

func TestUtilityUnitOfWork_GetAppMinimumPauseBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppMinimumPauseBlocks())
	gotParam, err := uow.getAppMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetAppMinimumStake(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetAppMinimumStake()
	gotParam, err := uow.getAppMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetAppUnstakingBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetAppUnstakingBlocks())
	gotParam, err := uow.getAppUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetBlocksPerSession(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetBlocksPerSession())
	gotParam, err := uow.getParameter(typesUtil.BlocksPerSessionParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetDoubleSignBurnPercentage(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetDoubleSignBurnPercentage())
	gotParam, err := uow.getDoubleSignBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetDoubleSignFeeOwner(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFeeOwner()
	gotParam, err := uow.getDoubleSignFeeOwner()
	require.NoError(t, err)

	defaultParamTx, er := hex.DecodeString(defaultParam)
	require.NoError(t, er)

	require.Equal(t, defaultParamTx, gotParam)
}

func TestUtilityUnitOfWork_GetFishermanMaxChains(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMaxChains())
	gotParam, err := uow.getFishermanMaxChains()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetFishermanMaxPausedBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMaxPauseBlocks())
	gotParam, err := uow.getFishermanMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetFishermanMinimumPauseBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMinimumPauseBlocks())
	gotParam, err := uow.getFishermanMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetFishermanMinimumStake(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetFishermanMinimumStake()
	gotParam, err := uow.getFishermanMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetFishermanUnstakingBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetFishermanUnstakingBlocks())
	gotParam, err := uow.getFishermanUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetMaxEvidenceAgeInBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaxEvidenceAgeInBlocks())
	gotParam, err := uow.getMaxEvidenceAgeInBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetMessageChangeParameterFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageChangeParameterFee()
	gotParam, err := uow.getMessageChangeParameterFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageDoubleSignFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFee()
	gotParam, err := uow.getMessageDoubleSignFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageEditStakeAppFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeAppFee()
	gotParam, err := uow.getMessageEditStakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageEditStakeFishermanFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeFishermanFee()
	gotParam, err := uow.getMessageEditStakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageEditStakeServicerFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeServicerFee()
	gotParam, err := uow.getMessageEditStakeServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageEditStakeValidatorFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeValidatorFee()
	gotParam, err := uow.getMessageEditStakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageFishermanPauseServicerFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageFishermanPauseServicerFee()
	gotParam, err := uow.getMessageFishermanPauseServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessagePauseAppFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseAppFee()
	gotParam, err := uow.getMessagePauseAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessagePauseFishermanFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseFishermanFee()
	gotParam, err := uow.getMessagePauseFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessagePauseServicerFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseServicerFee()
	gotParam, err := uow.getMessagePauseServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessagePauseValidatorFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseValidatorFee()
	gotParam, err := uow.getMessagePauseValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageProveTestScoreFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageProveTestScoreFee()
	gotParam, err := uow.getMessageProveTestScoreFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageSendFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageSendFee()
	gotParam, err := uow.getMessageSendFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageStakeAppFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeAppFee()
	gotParam, err := uow.getMessageStakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageStakeFishermanFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeFishermanFee()
	gotParam, err := uow.getMessageStakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageStakeServicerFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeServicerFee()
	gotParam, err := uow.getMessageStakeServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageStakeValidatorFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeValidatorFee()
	gotParam, err := uow.getMessageStakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageTestScoreFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageTestScoreFee()
	gotParam, err := uow.getMessageTestScoreFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnpauseAppFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseAppFee()
	gotParam, err := uow.getMessageUnpauseAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnpauseFishermanFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseFishermanFee()
	gotParam, err := uow.getMessageUnpauseFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnpauseServicerFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseServicerFee()
	gotParam, err := uow.getMessageUnpauseServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnpauseValidatorFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseValidatorFee()
	gotParam, err := uow.getMessageUnpauseValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnstakeAppFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeAppFee()
	gotParam, err := uow.getMessageUnstakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnstakeFishermanFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeFishermanFee()
	gotParam, err := uow.getMessageUnstakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnstakeServicerFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeServicerFee()
	gotParam, err := uow.getMessageUnstakeServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMessageUnstakeValidatorFee(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeValidatorFee()
	gotParam, err := uow.getMessageUnstakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetMissedBlocksBurnPercentage(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := uow.getMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetProposerPercentageOfFees(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetProposerPercentageOfFees())
	gotParam, err := uow.getProposerPercentageOfFees()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetServicerMaxChains(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServicerMaxChains())
	gotParam, err := uow.getServicerMaxChains()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetServicerMaxPausedBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServicerMaxPauseBlocks())
	gotParam, err := uow.getServicerMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetServicerMinimumPauseBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServicerMinimumPauseBlocks())
	gotParam, err := uow.getServicerMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetServicerMinimumStake(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetServicerMinimumStake()
	gotParam, err := uow.getServicerMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetServicerUnstakingBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetServicerUnstakingBlocks())
	gotParam, err := uow.getServicerUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetSessionTokensMultiplier(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppSessionTokensMultiplier())
	gotParam, err := uow.getAppSessionTokensMultiplier()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetValidatorMaxMissedBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaximumMissedBlocks())
	gotParam, err := uow.getValidatorMaxMissedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetValidatorMaxPausedBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaxPauseBlocks())
	gotParam, err := uow.getValidatorMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetValidatorMinimumPauseBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMinimumPauseBlocks())
	gotParam, err := uow.getValidatorMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_GetValidatorMinimumStake(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetValidatorMinimumStake()
	gotParam, err := uow.getValidatorMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, utils.BigIntToString(gotParam))
}

func TestUtilityUnitOfWork_GetValidatorUnstakingBlocks(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetValidatorUnstakingBlocks())
	gotParam, err := uow.getValidatorUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityUnitOfWork_HandleMessageChangeParameter(t *testing.T) {
	cdc := codec.GetCodec()
	uow := newTestingUtilityUnitOfWork(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := uow.getMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
	newParamValue := int32(2)
	paramOwnerPK := test_artifacts.DefaultParamsOwner
	any, er := cdc.ToAny(&wrapperspb.Int32Value{
		Value: newParamValue,
	})
	require.NoError(t, er)
	msg := &typesUtil.MessageChangeParameter{
		Owner:          paramOwnerPK.Address(),
		ParameterKey:   typesUtil.MissedBlocksBurnPercentageParamName,
		ParameterValue: any,
	}
	require.NoError(t, uow.handleMessageChangeParameter(msg), "handle message change param")
	gotParam, err = uow.getMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, int(newParamValue), gotParam)
}

func TestUtilityUnitOfWork_GetParamOwner(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	for _, paramName := range utils.GovParamMetadataKeys {
		paramOwnerName := utils.GovParamMetadataMap[paramName].ParamOwner
		paramOwner, err := uow.getParameter(paramOwnerName)
		require.NoError(t, err)
		paramOwnerString, ok := paramOwner.(string)
		require.True(t, ok)
		gotParam, err := uow.getParamOwner(paramName)
		require.NoError(t, err)
		require.Equal(t, paramOwnerString, hex.EncodeToString(gotParam))
	}
}
