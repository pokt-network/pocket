package test

import (
	"encoding/hex"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/utility"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_AddAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.AddAccountAmount(addrBz, addAmount), "add account amount")
	afterAmount, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
	// RESEARCH a golang specific solution for after test teardown
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_AddAccountAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addAmountString := types.BigIntToString(addAmount)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.AddAccountAmountString(addrBz, addAmountString), "add account amount string")
	afterAmount, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_AddPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(pool.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	require.NoError(t, ctx.AddPoolAmount(pool.GetAddress(), addAmount), "add pool amount")
	afterAmount, err := ctx.GetPoolAmount(pool.GetAddress())
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal")
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_HandleMessageSend(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	senderBalanceBefore, err := types.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceBefore, err := types.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)
	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	err = ctx.HandleMessageSend(&msg)
	require.NoError(t, err, "handle message send")

	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].GetAmount())
	require.NoError(t, err)

	recipientBalanceAfter, err := types.StringToBigInt(accs[1].GetAmount())
	require.NoError(t, err)
	require.Equal(t, sendAmount, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter))
	require.Equal(t, sendAmount, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore))
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	addrBz, er := hex.DecodeString(accs[0].GetAddress())
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].GetAddress())
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	candidates, err := ctx.GetMessageSendSignerCandidates(&msg)
	require.NoError(t, err)
	require.Equal(t, 1, len(candidates))
	require.Equal(t, addrBz, candidates[0])
	test_artifacts.CleanupTest(ctx)
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
	require.Equal(t, amount, gotAmountString)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	require.NoError(t, ctx.SetAccountAmount(addr, amount), "set account amount")
	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_SetAccountWithAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	amountString := types.BigIntToString(amount)
	require.NoError(t, ctx.SetAccountWithAmountString(addr, amountString), "set account amount string")
	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
	beforeAmount := pool.GetAmount()
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	expectedAfterAmount := big.NewInt(100)
	require.NoError(t, ctx.SetPoolAmount(pool.GetAddress(), expectedAfterAmount), "set pool amount")
	amount, err := ctx.GetPoolAmount(pool.GetAddress())
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)
	require.Equal(t, amount, expectedAfterAmount)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	beforeAmountBig := big.NewInt(1000000000000000)
	ctx.SetPoolAmount(pool.GetAddress(), beforeAmountBig)
	subAmountBig := big.NewInt(100)
	subAmount := types.BigIntToString(subAmountBig)
	require.NoError(t, ctx.SubPoolAmount(pool.GetAddress(), subAmount), "sub pool amount")
	amount, err := ctx.GetPoolAmount(pool.GetAddress())
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	beforeAmount := acc.GetAmount()
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	subAmountBig := big.NewInt(100)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.SubtractAccountAmount(addrBz, subAmountBig), "sub account amount")
	amount, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount)
	test_artifacts.CleanupTest(ctx)
}

func GetAllTestingAccounts(t *testing.T, ctx utility.UtilityContext) []*coreTypes.Account {
	accs, err := (ctx.Context.PersistenceRWContext).GetAllAccounts(0)
	require.NoError(t, err)
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].GetAddress() < accs[j].GetAddress()
	})
	return accs
}

func GetAllTestingPools(t *testing.T, ctx utility.UtilityContext) []*coreTypes.Account {
	accs, err := (ctx.Context.PersistenceRWContext).GetAllPools(0)
	require.NoError(t, err)
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].GetAddress() < accs[j].GetAddress()
	})
	return accs
}
