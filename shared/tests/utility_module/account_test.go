package utility_module

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility"
)

func TestUtilityContext_AddAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]
	initialAmount, err := types.StringToBigInt(acc.Amount)
	require.NoError(t, err)
	addAmount := big.NewInt(1)
	err = ctx.AddAccountAmount(acc.Address, addAmount)
	require.NoError(t, err, "add account amount")

	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)
	expected := initialAmount.Add(initialAmount, addAmount)
	require.True(t, afterAmount.Cmp(expected) == 0, fmt.Sprintf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount))
}

func TestUtilityContext_AddAccountAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]
	initialAmount, err := types.StringToBigInt(acc.Amount)
	require.NoError(t, err)
	addAmount := big.NewInt(1)
	addAmountString := types.BigIntToString(addAmount)
	err = ctx.AddAccountAmountString(acc.Address, addAmountString)
	require.NoError(t, err, "add account amount string")

	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)
	expected := initialAmount.Add(initialAmount, addAmount)
	require.True(t, afterAmount.Cmp(expected) == 0, fmt.Sprintf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount))
}

func TestUtilityContext_AddPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	initialAmount, err := types.StringToBigInt(pool.Account.Amount)
	require.NoError(t, err)
	addAmount := big.NewInt(1)
	err = ctx.AddPoolAmount(pool.Name, addAmount)
	require.NoError(t, err, "add pool amount")

	afterAmount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	expected := initialAmount.Add(initialAmount, addAmount)
	require.True(t, afterAmount.Cmp(expected) == 0, fmt.Sprintf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount))
}

func TestUtilityContext_HandleMessageSend(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)
	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	senderBalanceBefore, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)
	recipientBalanceBefore, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)
	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	err = ctx.HandleMessageSend(&msg)
	require.NoError(t, err, "handle message send")

	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)
	recipientBalanceAfter, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)
	require.True(t, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter).Cmp(sendAmount) == 0, fmt.Sprintf("unexpected sender balance"))
	require.True(t, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore).Cmp(sendAmount) == 0, fmt.Sprintf("unexpected recipient balance"))
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)
	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	candidates, err := ctx.GetMessageSendSignerCandidates(&msg)
	require.NoError(t, err)
	require.True(t, len(candidates) == 1, fmt.Sprintf("wrong number of candidates, expected %d, got %d", 1, len(candidates)))
	require.True(t, bytes.Equal(candidates[0], accs[0].Address), fmt.Sprintf("unexpected signer candidate"))
}

func TestUtilityContext_InsertPool(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	testPoolName := "TEST_POOL"
	addr, _ := crypto.GenerateAddress()
	amount := types.BigIntToString(big.NewInt(1000))
	err := ctx.InsertPool(testPoolName, addr, amount)
	require.NoError(t, err, "insert pool")

	gotAmount, err := ctx.GetPoolAmount(testPoolName)
	require.NoError(t, err)
	gotAmountString := types.BigIntToString(gotAmount)
	require.True(t, amount == gotAmountString, fmt.Sprintf("unexpected amount, expected %s got %s", amount, gotAmountString))
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	addr, _ := crypto.GenerateAddress()
	amount := big.NewInt(100)
	err := ctx.SetAccountAmount(addr, amount)
	require.NoError(t, err, "set account amount")

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.True(t, gotAmount.Cmp(amount) == 0, fmt.Sprintf("unexpected amounts: expected %v, got %v", amount, gotAmount))
}

func TestUtilityContext_SetAccountWithAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	addr, _ := crypto.GenerateAddress()
	amount := big.NewInt(100)
	amountString := types.BigIntToString(amount)
	err := ctx.SetAccountWithAmountString(addr, amountString)
	require.NoError(t, err, "set account amount string")

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.True(t, gotAmount.Cmp(amount) == 0, fmt.Sprintf("unexpected amounts: expected %v, got %v", amount, gotAmount))
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	beforeAmount := pool.Account.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)
	expectedAfterAmount := big.NewInt(100)
	err = ctx.SetPoolAmount(pool.Name, expectedAfterAmount)
	require.NoError(t, err, "set pool amount")

	amount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	require.True(t, beforeAmountBig.Cmp(amount) != 0, fmt.Sprintf("no amount change in pool"))
	require.True(t, expectedAfterAmount.Cmp(amount) == 0, fmt.Sprintf("unexpected pool amount; expected %v got %v", expectedAfterAmount, amount))
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	beforeAmountBig := big.NewInt(1000000000000000)
	ctx.SetPoolAmount(pool.Name, beforeAmountBig)
	subAmountBig := big.NewInt(100)
	subAmount := types.BigIntToString(subAmountBig)
	err := ctx.SubPoolAmount(pool.Name, subAmount)
	require.NoError(t, err, "sub pool amount")

	amount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	require.True(t, beforeAmountBig.Cmp(amount) != 0, fmt.Sprintf("no amount change in pool"))
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.True(t, expected.Cmp(amount) == 0, fmt.Sprintf("unexpected pool amount; expected %v got %v", expected, amount))
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]
	beforeAmount := acc.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)
	subAmountBig := big.NewInt(100)
	err = ctx.SubtractAccountAmount(acc.Address, subAmountBig)
	require.NoError(t, err, "sub account amount")

	amount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)
	require.True(t, beforeAmountBig.Cmp(amount) != 0, fmt.Sprintf("no amount change in pool"))
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.True(t, expected.Cmp(amount) == 0, fmt.Sprintf("unexpected acc amount; expected %v got %v", expected, amount))
}

func GetAllTestingAccounts(t *testing.T, ctx utility.UtilityContext) []*genesis.Account {
	accs, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllAccounts(0)
	require.NoError(t, err)
	return accs
}

func GetAllTestingPools(t *testing.T, ctx utility.UtilityContext) []*genesis.Pool {
	accs, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllPools(0)
	require.NoError(t, err)
	return accs
}
