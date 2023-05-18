package unit_of_work

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

var (
	defaultSendAmount = big.NewInt(10000)
)

func TestUtilityUnitOfWork_BasicValidateTransaction(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)

	tx, startingBalance, _, signer := newTestingTransaction(t, uow)
	msg, err := uow.basicValidateTransaction(tx)
	require.NoError(t, err)
	require.Equal(t, signer.Address().Bytes(), msg.GetSigner())
	feeBig, err := getGovParam[*big.Int](uow, typesUtil.MessageSendFee)
	require.NoError(t, err)

	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, feeBig)
	amount, err := uow.getAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amount, "unexpected after balance")
}

func TestUtilityUnitOfWork_ApplyTransaction(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)

	tx, startingBalance, amount, signer := newTestingTransaction(t, uow)
	idxTx, err := uow.HandleTransaction(tx, 0)
	require.NoError(t, err)
	require.Equal(t, int32(0), idxTx.GetResultCode())
	require.Equal(t, "", idxTx.GetError())
	feeBig, err := getGovParam[*big.Int](uow, typesUtil.MessageSendFee)
	require.NoError(t, err)

	expectedAmountSubtracted := amount.Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amount, err = uow.getAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amount, "unexpected after balance")
}

// TODO(@deblasis): refactor this to test HandleTransaction specifically in the utility package
// func TestUtilityUnitOfWork_HandleTransaction(t *testing.T) {
// 	ctx := newTestingUtilityContext(t, 0)
// 	tx, _, _, _ := newTestingTransaction(t, ctx)

// 	txBz, err := tx.Bytes()
// 	require.NoError(t, err)
// 	require.NoError(t, testUtilityMod.HandleTransaction(txBz))

// 	hash, err := tx.Hash()
// 	require.NoError(t, err)
// 	require.True(t, testUtilityMod.GetMempool().Contains(hash))
// 	require.Equal(t, testUtilityMod.HandleTransaction(txBz).Error(), typesUtil.ErrDuplicateTransaction().Error())
// }

func TestUtilityUnitOfWork_GetSignerCandidates(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	accs := getAllTestingAccounts(t, uow)

	sendAmount := big.NewInt(1000000)
	sendAmountString := utils.BigIntToString(sendAmount)
	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	candidates, err := uow.getSignerCandidates(&msg)
	require.NoError(t, err)

	require.Equal(t, 1, len(candidates), "wrong number of candidates")
	require.Equal(t, accs[0].GetAddress(), hex.EncodeToString(candidates[0]), "unexpected signer candidate")
}

// TODO(@deblasis): refactor this to test HandleTransaction specifically in the utility package
// func TestUtilityUnitOfWork_CreateAndApplyBlock(t *testing.T) {
// 	ctx := newTestingUtilityContext(t, 0)
// 	tx, _, _, _ := newTestingTransaction(t, ctx)

// 	proposer := getFirstActor(t, ctx, types.ActorType_ACTOR_TYPE_VAL)
// 	txBz, err := tx.Bytes()
// 	require.NoError(t, err)
// 	require.NoError(t, testUtilityMod.HandleTransaction(txBz))

// 	appHash, txs, er := ctx.CreateAndApplyProposalBlock([]byte(proposer.GetAddress()), 10000)
// 	require.NoError(t, er)
// 	require.NotEmpty(t, appHash)
// 	require.Equal(t, 1, len(txs))
// 	require.Equal(t, txs[0], txBz)
// }

func TestUtilityUnitOfWork_HandleMessage(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 0)
	accs := getAllTestingAccounts(t, uow)

	sendAmount := big.NewInt(1000000)
	sendAmountString := utils.BigIntToString(sendAmount)
	senderBalanceBefore, err := utils.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceBefore, err := utils.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)
	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	require.NoError(t, uow.handleMessageSend(&msg))
	accs = getAllTestingAccounts(t, uow)
	senderBalanceAfter, err := utils.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceAfter, err := utils.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)

	require.Equal(t, sendAmount, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter), "unexpected sender balance")
	require.Equal(t, sendAmount, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore), "unexpected recipient balance")
}

func newTestingTransaction(t *testing.T, uow *baseUtilityUnitOfWork) (tx *coreTypes.Transaction, startingBalance, amountSent *big.Int, signer crypto.PrivateKey) {
	amountSent = new(big.Int).Set(defaultSendAmount)
	startingBalance = new(big.Int).Set(test_artifacts.DefaultAccountAmount)

	recipientAddr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	signer, err = crypto.GeneratePrivateKey()
	require.NoError(t, err)

	signerAddr := signer.Address()
	require.NoError(t, uow.setAccountAmount(signerAddr, startingBalance))

	msg := NewTestingSendMessage(t, signerAddr, recipientAddr.Bytes(), utils.BigIntToString(amountSent))
	any, err := codec.GetCodec().ToAny(&msg)
	require.NoError(t, err)

	tx = &coreTypes.Transaction{
		Msg:   any,
		Nonce: testNonce,
	}
	require.NoError(t, tx.Sign(signer))

	return
}
