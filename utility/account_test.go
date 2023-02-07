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
	addrBz, err := hex.DecodeString(acc.GetAddress())
	require.NoError(t, err)
	require.NoError(t, ctx.addAccountAmount(addrBz, addAmount), "add account amount")

	afterAmount, err := ctx.getAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount)
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	acc := getFirstTestingAccount(t, ctx)

	beforeAmount, err := converters.StringToBigInt(acc.GetAmount())
	require.NoError(t, err)

	subAmount := big.NewInt(100)
	addrBz, er := hex.DecodeString(acc.GetAddress())
	require.NoError(t, er)
	require.NoError(t, ctx.subtractAccountAmount(addrBz, subAmount), "sub account amount")

	amount, err := ctx.getAccountAmount(addrBz)
	require.NoError(t, err)
	require.NotEqual(t, beforeAmount, amount)

	expected := beforeAmount.Sub(beforeAmount, subAmount)
	require.Equal(t, expected, amount)
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

	amount := big.NewInt(1000)
	err = ctx.insertPool(testPoolName, addr, amount)
	require.NoError(t, err, "insert pool")

	poolAmount, err := ctx.getPoolAmount(testPoolName)
	require.NoError(t, err)
	require.Equal(t, amount, poolAmount)
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	pool := getFirstTestingPool(t, ctx)

	beforeAmount, err := converters.StringToBigInt(pool.GetAmount())
	require.NoError(t, err)

	expectedAfterAmount := big.NewInt(100)
	require.NoError(t, ctx.setPoolAmount(pool.GetAddress(), expectedAfterAmount), "set pool amount")

	amount, err := ctx.getPoolAmount(pool.GetAddress())
	require.NoError(t, err)
	require.NotEqual(t, beforeAmount, amount)
	require.Equal(t, amount, expectedAfterAmount)
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	pool := getFirstTestingPool(t, ctx)

	beforeAmountBig := big.NewInt(1000000000000000)
	ctx.setPoolAmount(pool.GetAddress(), beforeAmountBig)
	subAmount := big.NewInt(100)
	require.NoError(t, ctx.subPoolAmount(pool.GetAddress(), subAmount), "sub pool amount")

	amount, err := ctx.getPoolAmount(pool.GetAddress())
	require.NoError(t, err)
	require.NotEqual(t, beforeAmountBig, amount)

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmount)
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
