package utility_module

import (
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
	require.Equal(t, afterAmount, expected, "amounts are not equal")
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
	require.Equal(t, afterAmount, expected, "amounts are not equal")
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
	require.Equal(t, afterAmount, expected, "amounts are not equal")
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
	require.Equal(t, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter), sendAmount, "unexpected sender balance")
	require.Equal(t, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore), sendAmount, "unexpected recipient balance")
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)

	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	candidates, err := ctx.GetMessageSendSignerCandidates(&msg)
	require.NoError(t, err)
	require.Equal(t, 1, len(candidates), "wrong number of candidates")
	require.Equal(t, candidates[0], accs[0].Address, "unexpected signer candidate")
}

func TestUtilityContext_InsertPool(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	testPoolName := "TEST_POOL"

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := types.BigIntToString(big.NewInt(1000))
	err = ctx.InsertPool(testPoolName, addr, amount)
	require.NoError(t, err, "insert pool")

	gotAmount, err := ctx.GetPoolAmount(testPoolName)
	require.NoError(t, err)

	gotAmountString := types.BigIntToString(gotAmount)
	require.True(t, amount == gotAmountString, fmt.Sprintf("unexpected amount"))
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	err = ctx.SetAccountAmount(addr, amount)
	require.NoError(t, err, "set account amount")

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, gotAmount, amount, "unexpected amounts")
}

func TestUtilityContext_SetAccountWithAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	amountString := types.BigIntToString(amount)
	err = ctx.SetAccountWithAmountString(addr, amountString)
	require.NoError(t, err, "set account amount string")

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, gotAmount, amount, "unexpected amounts: expected")
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
	require.NotEqual(t, beforeAmountBig, amount, "no amount change in pool")
	require.Equal(t, expectedAfterAmount, amount, "unexpected pool amount")
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
	require.NotEqual(t, beforeAmountBig, amount, "no amount change in pool")

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount, "unexpected pool amount")
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
	require.NotEqual(t, beforeAmountBig, amount, "no amount change in pool")

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount, "unexpected acc amount")
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
