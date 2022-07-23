package utility_module

import (
	"bytes"
	"encoding/hex"
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
	if _, err := ctx.AnteHandleMessage(tx); err != nil {
		t.Fatal(err)
	}
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, feeBig)
	amount, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	if amount.Cmp(expectedAfterBalance) != 0 {
		t.Fatalf("unexpected after balance; expected %v got %v", expectedAfterBalance, amount)
	}
}

func TestUtilityContext_ApplyTransaction(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, startingBalance, amount, signer := NewTestingTransaction(t, ctx)
	if err := ctx.ApplyTransaction(tx); err != nil {
		t.Fatal(err)
	}
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	expectedAmountSubtracted := amount.Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amount, err = ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	if amount.Cmp(expectedAfterBalance) != 0 {
		t.Fatalf("unexpected after balance; expected %v got %v", expectedAfterBalance, amount)
	}
}

func TestUtilityContext_CheckTransaction(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := NewTestingTransaction(t, ctx)
	txBz, err := tx.Bytes()
	require.NoError(t, err)
	if err := ctx.CheckTransaction(txBz); err != nil {
		t.Fatal(err)
	}
	hash, err := tx.Hash()
	require.NoError(t, err)
	if !ctx.Mempool.Contains(hash) {
		t.Fatal("the transaction was unable to be checked")
	}
	if err := ctx.CheckTransaction(txBz); err.Error() != types.ErrDuplicateTransaction().Error() {
		t.Fatalf("unexpected err, expected %v got %v", types.ErrDuplicateTransaction().Error(), err.Error())
	}
}

func TestUtilityContext_GetSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)
	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	candidates, err := ctx.GetSignerCandidates(&msg)
	require.NoError(t, err)
	if len(candidates) != 1 {
		t.Fatalf("wrong number of candidates, expected %d, got %d", 1, len(candidates))
	}
	if !bytes.Equal(candidates[0], accs[0].Address) {
		t.Fatal("unexpected signer candidate")
	}
}

func TestUtilityContext_GetProposalTransactions(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := NewTestingTransaction(t, ctx)
	proposer := GetAllTestingValidators(t, ctx)[0]
	txBz, err := tx.Bytes()
	require.NoError(t, err)
	if err := ctx.CheckTransaction(txBz); err != nil {
		t.Fatal(err)
	}
	txs, er := ctx.GetProposalTransactions(proposer.Address, 10000, nil)
	if er != nil {
		t.Fatal(er)
	}
	if len(txs) != 1 {
		t.Fatalf("incorrect txs amount returned; expected %v got %v", 1, len(txs))
	}
	if !bytes.Equal(txs[0], txBz) {
		t.Fatalf("unexpected transaction returned; expected tx: %s, got %s", hex.EncodeToString(txBz), hex.EncodeToString(txs[0]))
	}
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
	if err := ctx.HandleMessageSend(&msg); err != nil {
		t.Fatal(err)
	}
	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)
	recipientBalanceAfter, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)
	if big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter).Cmp(sendAmount) != 0 {
		t.Fatal("unexpected sender balance")
	}
	if big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore).Cmp(sendAmount) != 0 {
		t.Fatal("unexpected recipient balance")
	}
}

func NewTestingTransaction(t *testing.T, ctx utility.UtilityContext) (transaction *typesUtil.Transaction, startingAmount, amountSent *big.Int, signer crypto.PrivateKey) {
	var err error
	cdc := types.GetCodec()
	recipient := GetAllTestingAccounts(t, ctx)[1]
	signer, err = crypto.GeneratePrivateKey()
	require.NoError(t, err)
	startingAmount = defaultAmount
	signerAddr := signer.Address()
	if err = ctx.SetAccountAmount(signerAddr, defaultAmount); err != nil {
		t.Fatal(err)
	}
	amountSent = defaultSendAmount
	msg := NewTestingSendMessage(t, signerAddr, recipient.Address, defaultSendAmountString)
	any, err := cdc.ToAny(&msg)
	require.NoError(t, err)
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	fee := types.BigIntToString(feeBig)
	transaction = &typesUtil.Transaction{
		Msg:   any,
		Fee:   fee,
		Nonce: defaultNonceString,
	}
	if err = transaction.Sign(signer); err != nil {
		t.Fatal(err)
	}
	return
}
