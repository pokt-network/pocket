package utility_module

import (
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_AnteHandleMessage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	tx, startingBalance, _, signer := NewTestingTransaction(t, ctx)
	_, err := ctx.AnteHandleMessage(tx)
	require.NoError(t, err)

	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, feeBig)
	amount, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, amount, expectedAfterBalance, "unexpected after balance")
}

func TestUtilityContext_ApplyTransaction(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	tx, startingBalance, amount, signer := NewTestingTransaction(t, ctx)
	require.NoError(t, ctx.ApplyTransaction(tx))

	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	expectedAmountSubtracted := amount.Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amount, err = ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, amount, expectedAfterBalance, "unexpected after balance")
}

func TestUtilityContext_CheckTransaction(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	tx, _, _, _ := NewTestingTransaction(t, ctx)
	txBz, err := tx.Bytes()
	require.NoError(t, err)
	require.NoError(t, ctx.CheckTransaction(txBz))

	hash, err := tx.Hash()
	require.NoError(t, err)
	require.True(t, ctx.Mempool.Contains(hash), "the transaction was unable to be checked")

	er := ctx.CheckTransaction(txBz)
	require.Equal(t, er.Error(), types.ErrDuplicateTransaction().Error())
}

func TestUtilityContext_GetSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	candidates, err := ctx.GetSignerCandidates(&msg)
	require.NoError(t, err)
	require.Equal(t, len(candidates), 1, "wrong number of candidates")
	require.Equal(t, candidates[0], accs[0].Address, "unexpected signer candidate")
}

func TestUtilityContext_GetProposalTransactions(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := NewTestingTransaction(t, ctx)
	proposer := GetAllTestingValidators(t, ctx)[0]

	txBz, err := tx.Bytes()
	require.NoError(t, err)
	require.NoError(t, ctx.CheckTransaction(txBz))

	txs, er := ctx.GetTransactionsForProposal(proposer.Address, 10000, nil)
	require.NoError(t, er)
	require.Equal(t, len(txs), 1, "incorrect txs amount returned")
	require.Equal(t, txs[0], txBz, "unexpected transaction returned")
}

func TestUtilityContext_HandleMessage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	senderBalanceBefore, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)

	recipientBalanceBefore, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)

	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	require.NoError(t, ctx.HandleMessageSend(&msg))

	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)

	recipientBalanceAfter, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter), sendAmount, "unexpected sender balance")
	require.Equal(t, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore), sendAmount, "unexpected recipient balance")
}

func NewTestingTransaction(t *testing.T, ctx utility.UtilityContext) (transaction *typesUtil.Transaction, startingAmount, amountSent *big.Int, signer crypto.PrivateKey) {
	cdc := types.GetCodec()
	recipient := GetAllTestingAccounts(t, ctx)[1]

	signer, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	startingAmount = defaultAmount
	signerAddr := signer.Address()
	require.NoError(t, ctx.SetAccountAmount(signerAddr, defaultAmount))

	amountSent = defaultSendAmount
	msg := NewTestingSendMessage(t, signerAddr, recipient.Address, defaultSendAmountString)
	any, err := cdc.ToAny(&msg)
	require.NoError(t, err)

	transaction = &typesUtil.Transaction{
		Msg:   any,
		Nonce: defaultNonceString,
	}
	require.NoError(t, transaction.Sign(signer))
	return
}
