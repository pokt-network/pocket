package utility

import (
	"encoding/hex"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_AddAccountAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	acc := getFirstTestingAccount(t, ctx)

	initialAmount, err := converters.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.addAccountAmount(addrBz, addAmount), "add account amount")

	afterAmount, err := ctx.getAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
}

func TestUtilityContext_AddAccountAmountString(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	acc := getFirstTestingAccount(t, ctx)

	initialAmount, err := converters.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addAmountString := converters.BigIntToString(addAmount)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.addAccountAmountString(addrBz, addAmountString), "add account amount string")

	afterAmount, err := ctx.getAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	require.NoError(t, ctx.setAccountAmount(addr, amount), "set account amount")
	gotAmount, err := ctx.getAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount)
}

func TestUtilityContext_SetAccountWithAmountString(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	amountString := converters.BigIntToString(amount)
	require.NoError(t, ctx.setAccountWithAmountString(addr, amountString), "set account amount string")

	gotAmount, err := ctx.getAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount)
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	acc := getFirstTestingAccount(t, ctx)

	beforeAmount := acc.GetAmount()
	beforeAmountBig, err := converters.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	subAmountBig := big.NewInt(100)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.subtractAccountAmount(addrBz, subAmountBig), "sub account amount")

	amount, err := ctx.getAccountAmount(addrBz)
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount)
}

func TestUtilityContext_AddPoolAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	pool := getFirstTestingPool(t, ctx)

	initialAmount, err := converters.StringToBigInt(pool.GetAmount())
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	require.NoError(t, ctx.addPoolAmount(pool.GetAddress(), addAmount), "add pool amount")

	afterAmount, err := ctx.getPoolAmount(pool.GetAddress())
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
}

func TestUtilityContext_InsertPool(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	testPoolName := "TEST_POOL"

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := converters.BigIntToString(big.NewInt(1000))
	err = ctx.insertPool(testPoolName, addr, amount)
	require.NoError(t, err, "insert pool")

	poolAmount, err := ctx.getPoolAmount(testPoolName)
	require.NoError(t, err)

	poolAmountString := converters.BigIntToString(poolAmount)
	require.Equal(t, amount, poolAmountString)
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	pool := getFirstTestingPool(t, ctx)

	beforeAmount := pool.GetAmount()
	beforeAmountBig, err := converters.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	expectedAfterAmount := big.NewInt(100)
	require.NoError(t, ctx.setPoolAmount(pool.GetAddress(), expectedAfterAmount), "set pool amount")

	amount, err := ctx.getPoolAmount(pool.GetAddress())
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)
	require.Equal(t, amount, expectedAfterAmount)
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	pool := getFirstTestingPool(t, ctx)

	beforeAmountBig := big.NewInt(1000000000000000)
	ctx.setPoolAmount(pool.GetAddress(), beforeAmountBig)
	subAmountBig := big.NewInt(100)
	subAmount := converters.BigIntToString(subAmountBig)
	require.NoError(t, ctx.subPoolAmount(pool.GetAddress(), subAmount), "sub pool amount")

	amount, err := ctx.getPoolAmount(pool.GetAddress())
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount)
}

func getAllTestingAccounts(t *testing.T, ctx *utilityContext) []*coreTypes.Account {
	accs, err := (ctx.persistenceContext).GetAllAccounts(0)
	require.NoError(t, err)
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].GetAddress() < accs[j].GetAddress()
	})
	return accs
}

func getFirstTestingAccount(t *testing.T, ctx *utilityContext) *coreTypes.Account {
	return getAllTestingAccounts(t, ctx)[0]
}

func getAllTestingPools(t *testing.T, ctx *utilityContext) []*coreTypes.Account {
	pools, err := (ctx.persistenceContext).GetAllPools(0)
	require.NoError(t, err)
	sort.Slice(pools, func(i, j int) bool {
		return pools[i].GetAddress() < pools[j].GetAddress()
	})
	return pools
}

func getFirstTestingPool(t *testing.T, ctx *utilityContext) *coreTypes.Account {
	return getAllTestingPools(t, ctx)[0]
}
