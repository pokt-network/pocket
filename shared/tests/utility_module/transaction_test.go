package utility_module

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/tests"

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
	require.Equal(t, expectedAfterBalance, amount, "unexpected after balance")

	ctx.Context.Release()
	tests.CleanupTest()
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
	require.Equal(t, expectedAfterBalance, amount, "unexpected after balance")

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_CheckTransaction(t *testing.T) {
	//ctx := NewTestingUtilityContext(t, 0) TODO (Team) txIndexer not implemented by postgres context
	//tx, _, _, _ := NewTestingTransaction(t, ctx)
	//txBz, err := tx.Bytes()
	//require.NoError(t, err)
	//require.NoError(t, ctx.CheckTransaction(txBz))
	//hash, err := tx.Hash()
	//require.NoError(t, err)
	//require.True(t, ctx.Mempool.Contains(hash), fmt.Sprintf("the transaction was unable to be checked"))
	//er := ctx.CheckTransaction(txBz)
	//require.True(t, er.Error() == types.ErrDuplicateTransaction().Error(), fmt.Sprintf("unexpected err, expected %v got %v", types.ErrDuplicateTransaction().Error(), er.Error()))

	// ctx.Context.Release()
	// tests.CleanupTest()
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
	require.Equal(t, bytes.Equal(candidates[0], accs[0].Address), "unexpected signer candidate")

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_GetTransactionsForProposal(t *testing.T) {
	//ctx := NewTestingUtilityContext(t, 0) TODO (Team) txIndexer not implemented by postgres context
	//tx, _, _, _ := NewTestingTransaction(t, ctx)
	//proposer := GetAllTestingValidators(t, ctx)[0]
	//txBz, err := tx.Bytes()
	//require.NoError(t, err)
	//require.NoError(t, ctx.CheckTransaction(txBz))
	//txs, er := ctx.GetTransactionsForProposal(proposer.Address, 10000, nil)
	//require.NoError(t, er)
	//require.True(t, len(txs) == 1, fmt.Sprintf("incorrect txs amount returned; expected %v got %v", 1, len(txs)))
	//require.True(t, bytes.Equal(txs[0], txBz), fmt.Sprintf("unexpected transaction returned; expected tx: %s, got %s", hex.EncodeToString(txBz), hex.EncodeToString(txs[0])))

	// ctx.Context.Release()
	// tests.CleanupTest()
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

	ctx.Context.Release()
	tests.CleanupTest()
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
