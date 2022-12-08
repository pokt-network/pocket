package test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	utilTypes "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

var (
	defaultSendAmount = big.NewInt(10000)
)

func TestUtilityContext_AnteHandleMessage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	tx, startingBalance, _, signer := newTestingTransaction(t, ctx)
	_, signerString, err := ctx.AnteHandleMessage(tx)
	require.NoError(t, err)
	require.Equal(t, signer.Address().String(), signerString)
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, feeBig)
	amount, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amount, "unexpected after balance")

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_ApplyTransaction(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	tx, startingBalance, amount, signer := newTestingTransaction(t, ctx)
	txResult, err := ctx.ApplyTransaction(0, tx)
	require.NoError(t, err)
	require.Equal(t, int32(0), txResult.GetResultCode())
	require.Equal(t, "", txResult.GetError())
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	expectedAmountSubtracted := amount.Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amount, err = ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amount, "unexpected after balance")

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_CheckTransaction(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)
	txBz, err := tx.Bytes()
	require.NoError(t, err)
	require.NoError(t, ctx.CheckTransaction(txBz))
	hash, err := tx.Hash()
	require.NoError(t, err)
	require.True(t, ctx.Mempool.Contains(hash))
	er := ctx.CheckTransaction(txBz)
	require.Equal(t, er.Error(), typesUtil.ErrDuplicateTransaction().Error())

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := typesUtil.BigIntToString(sendAmount)
	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	candidates, err := ctx.GetSignerCandidates(&msg)
	require.NoError(t, err)

	require.Equal(t, 1, len(candidates), "wrong number of candidates")
	require.Equal(t, accs[0].GetAddress(), hex.EncodeToString(candidates[0]), "unexpected signer candidate")

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_CreateAndApplyBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)
	proposer := getAllTestingValidators(t, ctx)[0]
	txBz, err := tx.Bytes()
	require.NoError(t, err)
	require.NoError(t, ctx.CheckTransaction(txBz))
	appHash, txs, er := ctx.CreateAndApplyProposalBlock([]byte(proposer.GetAddress()), 10000)
	require.NoError(t, er)
	require.Equal(t, len(txs), 1)
	require.Equal(t, txs[0], txBz)
	require.NotEmpty(t, appHash)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_HandleMessage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := typesUtil.BigIntToString(sendAmount)
	senderBalanceBefore, err := typesUtil.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceBefore, err := typesUtil.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)
	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	require.NoError(t, ctx.HandleMessageSend(&msg))
	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := typesUtil.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceAfter, err := typesUtil.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)

	require.Equal(t, sendAmount, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter), "unexpected sender balance")
	require.Equal(t, sendAmount, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore), "unexpected recipient balance")

	test_artifacts.CleanupTest(ctx)
}

func newTestingTransaction(t *testing.T, ctx utility.UtilityContext) (transaction *typesUtil.Transaction, startingBalance, amountSent *big.Int, signer crypto.PrivateKey) {
	amountSent = new(big.Int).Set(defaultSendAmount)
	startingBalance = new(big.Int).Set(defaults.DefaultAccountAmount)

	recipientAddr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	signer, err = crypto.GeneratePrivateKey()
	require.NoError(t, err)

	signerAddr := signer.Address()
	require.NoError(t, ctx.SetAccountAmount(signerAddr, startingBalance))

	msg := NewTestingSendMessage(t, signerAddr, recipientAddr.Bytes(), utilTypes.BigIntToString(amountSent))
	any, err := codec.GetCodec().ToAny(&msg)
	require.NoError(t, err)

	transaction = &typesUtil.Transaction{
		Msg:   any,
		Nonce: testNonce,
	}
	require.NoError(t, transaction.Sign(signer))

	return
}
