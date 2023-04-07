package unit_of_work

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	utilTypes "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityUnitOfWork_ApplyBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUtilityMod := mockModules.NewMockUtilityModule(ctrl)
	mockUtilityMod.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()
	mockUtilityMod.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()

	uow := newTestingUtilityUnitOfWork(t, 0, func(uow *baseUtilityUnitOfWork) {
		uow.GetBus().RegisterModule(mockUtilityMod)
		mockUtilityMod.EXPECT().GetMempool().Return(NewTestingMempool(t)).AnyTimes()
	})
	tx, startingBalance, amountSent, signer := newTestingTransaction(t, uow)

	txBz, er := tx.Bytes()
	require.NoError(t, er)

	proposer := getFirstActor(t, uow, coreTypes.ActorType_ACTOR_TYPE_VAL)

	addrBz, err := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, err)

	proposerBeforeBalance, err := uow.getAccountAmount(addrBz)
	require.NoError(t, err)

	// calling ApplyBlock without having called SetProposalBlock first should fail with ErrProposalBlockNotSet
	err = uow.ApplyBlock()
	require.Equal(t, err.Error(), utilTypes.ErrProposalBlockNotSet().Error())

	err = uow.SetProposalBlock(IgnoreProposalBlockCheckHash, addrBz, [][]byte{txBz})
	require.NoError(t, err)

	err = uow.ApplyBlock()
	stateHash := uow.GetStateHash()
	require.NoError(t, err)
	require.NotNil(t, stateHash)

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.getValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.Equal(t, missed, 1)

	feeBig, err := getGovParam[*big.Int](uow, utilTypes.MessageSendFee)
	require.NoError(t, err)

	expectedAmountSubtracted := big.NewInt(0).Add(amountSent, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amountAfter, err := uow.getAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amountAfter, "unexpected after balance; expected %v got %v", expectedAfterBalance, amountAfter)

	proposerCutPercentage, err := getGovParam[int](uow, utilTypes.ProposerPercentageOfFeesParamName)
	require.NoError(t, err)

	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)

	proposerAfterBalance, err := uow.getAccountAmount(addrBz)
	require.NoError(t, err)

	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	require.Equal(t, expectedProposerBalanceDifference, proposerBalanceDifference, "unexpected before / after balance difference")
}

func TestUtilityUnitOfWork_BeginBlock(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	tx, _, _, _ := newTestingTransaction(t, uow)

	proposer := getFirstActor(t, uow, coreTypes.ActorType_ACTOR_TYPE_VAL)

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)

	er = uow.SetProposalBlock(IgnoreProposalBlockCheckHash, addrBz, [][]byte{txBz})
	require.NoError(t, er)

	er = uow.ApplyBlock()
	require.NoError(t, er)

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.getValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.Equal(t, missed, 1)
}

func TestUtilityUnitOfWork_EndBlock(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	tx, _, _, _ := newTestingTransaction(t, uow)

	proposer := getFirstActor(t, uow, coreTypes.ActorType_ACTOR_TYPE_VAL)

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)

	proposerBeforeBalance, err := uow.getAccountAmount(addrBz)
	require.NoError(t, err)

	er = uow.SetProposalBlock(IgnoreProposalBlockCheckHash, addrBz, [][]byte{txBz})
	require.NoError(t, er)

	er = uow.ApplyBlock()
	require.NoError(t, er)

	feeBig, err := getGovParam[*big.Int](uow, utilTypes.MessageSendFee)
	require.NoError(t, err)

	proposerCutPercentage, err := getGovParam[int](uow, utilTypes.ProposerPercentageOfFeesParamName)
	require.NoError(t, err)

	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := uow.getAccountAmount(addrBz)
	require.NoError(t, err)

	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	require.Equal(t, expectedProposerBalanceDifference, proposerBalanceDifference)
}
