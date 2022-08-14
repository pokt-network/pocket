package utility_module

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
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
	require.True(t, amount.Cmp(expectedAfterBalance) == 0, fmt.Sprintf("unexpected after balance; expected %v got %v", expectedAfterBalance, amount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
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
	require.True(t, amount.Cmp(expectedAfterBalance) == 0, fmt.Sprintf("unexpected after balance; expected %v got %v", expectedAfterBalance, amount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
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
	//ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
}

func TestUtilityContext_GetSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	addrBz, er := hex.DecodeString(accs[0].Address)
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].Address)
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	candidates, err := ctx.GetSignerCandidates(&msg)
	require.NoError(t, err)

	require.True(t, len(candidates) == 1, fmt.Sprintf("wrong number of candidates, expected %d, got %d", 1, len(candidates)))
	require.True(t, bytes.Equal(candidates[0], addrBz), fmt.Sprintf("unexpected signer candidate"))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
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
	//ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
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
	addrBz, er := hex.DecodeString(accs[0].Address)
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].Address)
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	require.NoError(t, ctx.HandleMessageSend(&msg))
	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)

	recipientBalanceAfter, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)

	require.True(t, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter).Cmp(sendAmount) == 0, fmt.Sprintf("unexpected sender balance"))
	require.True(t, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore).Cmp(sendAmount) == 0, fmt.Sprintf("unexpected recipient balance"))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
}

func NewTestingTransaction(t *testing.T, ctx utility.UtilityContext) (transaction *typesUtil.Transaction, startingAmount, amountSent *big.Int, signer crypto.PrivateKey) {
	cdc := types.GetCodec()
	recipient := GetAllTestingAccounts(t, ctx)[1]

	signer, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	startingAmount = test_artifacts.DefaultAccountAmount
	signerAddr := signer.Address()
	require.NoError(t, ctx.SetAccountAmount(signerAddr, test_artifacts.DefaultAccountAmount))
	amountSent = defaultSendAmount
	addrBz, err := hex.DecodeString(recipient.Address)
	require.NoError(t, err)
	msg := NewTestingSendMessage(t, signerAddr, addrBz, defaultSendAmountString)
	any, err := cdc.ToAny(&msg)
	require.NoError(t, err)
	transaction = &typesUtil.Transaction{
		Msg:   any,
		Nonce: defaultNonceString,
	}
	require.NoError(t, transaction.Sign(signer))
	return
}
