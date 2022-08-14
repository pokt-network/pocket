package utility_module

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/tests"

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
	require.NoError(t, ctx.AddAccountAmount(acc.Address, addAmount), "add account amount")

	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal, expected %v, got %v", initialAmount, afterAmount)

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_AddAccountAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(acc.Amount)
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addAmountString := types.BigIntToString(addAmount)
	require.NoError(t, ctx.AddAccountAmountString(acc.Address, addAmountString), "add account amount string")

	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal, expected %v, got %v", initialAmount, afterAmount)

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_AddPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(pool.Account.Amount)
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	require.NoError(t, ctx.AddPoolAmount(pool.Name, addAmount), "add pool amount")

	afterAmount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal")

	ctx.Context.Release()
	tests.CleanupTest()
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
	require.Equal(t, sendAmount, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter), "unexpected sender balance")
	require.Equal(t, sendAmount, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore), "unexpected recipient balance")

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)

	msg := NewTestingSendMessage(t, accs[0].Address, accs[1].Address, sendAmountString)
	candidates, err := ctx.GetMessageSendSignerCandidates(&msg)
	require.NoError(t, err)
	require.Equal(t, len(candidates), 1, fmt.Sprintf("wrong number of candidates, expected %d, got %d", 1, len(candidates)))
	require.Equal(t, candidates[0], accs[0].Address, "unexpected signer candidate")

	ctx.Context.Release()
	tests.CleanupTest()
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
	require.Equal(t, amount, gotAmountString, "unexpected amount")

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	require.NoError(t, ctx.SetAccountAmount(addr, amount), "set account amount")

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount, "unexpected amounts: expected %v, got %v", amount, gotAmount)

	ctx.Context.Release()
	tests.CleanupTest()
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
	require.Equal(t, amount, gotAmount, "unexpected amounts: expected %v, got %v", amount, gotAmount)

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	beforeAmount := pool.Account.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	expectedAfterAmount := big.NewInt(100)
	require.NoError(t, ctx.SetPoolAmount(pool.Name, expectedAfterAmount), "set pool amount")

	amount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	require.Equal(t, beforeAmountBig, defaultAmount, "no amount change in pool")
	require.Equal(t, expectedAfterAmount, amount, "unexpected pool amount; expected %v got %v", expectedAfterAmount, amount)

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	beforeAmountBig := big.NewInt(1000000000000000)
	ctx.SetPoolAmount(pool.Name, beforeAmountBig)

	subAmountBig := big.NewInt(100)
	subAmount := types.BigIntToString(subAmountBig)
	require.NoError(t, ctx.SubPoolAmount(pool.Name, subAmount), "sub pool amount")

	amount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, amount, expected, "unexpected pool amount; expected %v got %v", expected, amount)

	ctx.Context.Release()
	tests.CleanupTest()
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	beforeAmount := acc.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	subAmountBig := big.NewInt(100)
	require.NoError(t, ctx.SubtractAccountAmount(acc.Address, subAmountBig), "sub account amount")

	amount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)
	require.Equal(t, beforeAmountBig, defaultAmount, "no amount change in pool")

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount, "unexpected acc amount; expected %v got %v", expected, amount)

	ctx.Context.Release()
	tests.CleanupTest()
}

func GetAllTestingAccounts(t *testing.T, ctx utility.UtilityContext) []*genesis.Account {
	accs, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllAccounts(0)
	sort.Slice(accs, func(i, j int) bool {
		return hex.EncodeToString(accs[i].Address) < hex.EncodeToString(accs[j].Address)
	})
	require.NoError(t, err)
	return accs
}

func GetAllTestingPools(t *testing.T, ctx utility.UtilityContext) []*genesis.Pool {
	accs, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllPools(0)
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].Name < accs[j].Name
	})
	require.NoError(t, err)
	return accs
}
