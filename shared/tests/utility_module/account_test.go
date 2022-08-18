package utility_module

import (
<<<<<<< HEAD
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types/genesis"
=======
	"encoding/hex"
>>>>>>> main
	"math/big"
	"sort"
	"testing"

<<<<<<< HEAD
	"github.com/stretchr/testify/require"

=======
	"github.com/pokt-network/pocket/persistence"
>>>>>>> main
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/utility"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_AddAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(acc.Amount)
	require.NoError(t, err)

	addAmount := big.NewInt(1)
<<<<<<< HEAD
	addrBz, er := hex.DecodeString(acc.Address)
	require.NoError(t, er)
	require.NoError(t, ctx.AddAccountAmount(addrBz, addAmount), "add account amount")
	afterAmount, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.True(t, afterAmount.Cmp(expected) == 0, fmt.Sprintf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()   // TODO (team) need a golang specific solution for teardown
=======
	require.NoError(t, ctx.AddAccountAmount(acc.Address, addAmount), "add account amount")
	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_AddAccountAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(acc.Amount)
	require.NoError(t, err)

	addAmount := big.NewInt(1)
	addAmountString := types.BigIntToString(addAmount)
<<<<<<< HEAD
	addrBz, er := hex.DecodeString(acc.Address)
	require.NoError(t, er)
	require.NoError(t, ctx.AddAccountAmountString(addrBz, addAmountString), "add account amount string")
	afterAmount, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.True(t, afterAmount.Cmp(expected) == 0, fmt.Sprintf("amounts are not equal, expected %v, got %v", initialAmount, afterAmount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.NoError(t, ctx.AddAccountAmountString(acc.Address, addAmountString), "add account amount string")
	afterAmount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_AddPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	initialAmount, err := types.StringToBigInt(pool.Amount)
	require.NoError(t, err)

	addAmount := big.NewInt(1)
<<<<<<< HEAD
	require.NoError(t, ctx.AddPoolAmount(pool.Address, addAmount), "add pool amount")
	afterAmount, err := ctx.GetPoolAmount(pool.Address)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, afterAmount, expected, "amounts are not equal")
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.NoError(t, ctx.AddPoolAmount(pool.Name, addAmount), "add pool amount")
	afterAmount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)

	expected := initialAmount.Add(initialAmount, addAmount)
	require.Equal(t, expected, afterAmount, "amounts are not equal")

	tests.CleanupTest(ctx)
>>>>>>> main
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
	addrBz, er := hex.DecodeString(accs[0].Address)
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].Address)
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	err = ctx.HandleMessageSend(&msg)
	require.NoError(t, err, "handle message send")

	accs = GetAllTestingAccounts(t, ctx)
	senderBalanceAfter, err := types.StringToBigInt(accs[0].Amount)
	require.NoError(t, err)

	recipientBalanceAfter, err := types.StringToBigInt(accs[1].Amount)
	require.NoError(t, err)
<<<<<<< HEAD
	require.True(t, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter).Cmp(sendAmount) == 0, fmt.Sprintf("unexpected sender balance"))
	require.True(t, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore).Cmp(sendAmount) == 0, fmt.Sprintf("unexpected recipient balance"))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.Equal(t, sendAmount, big.NewInt(0).Sub(senderBalanceBefore, senderBalanceAfter), "unexpected sender balance")
	require.Equal(t, sendAmount, big.NewInt(0).Sub(recipientBalanceAfter, recipientBalanceBefore), "unexpected recipient balance")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_GetMessageSendSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	accs := GetAllTestingAccounts(t, ctx)

	sendAmount := big.NewInt(1000000)
	sendAmountString := types.BigIntToString(sendAmount)
	addrBz, er := hex.DecodeString(accs[0].Address)
	require.NoError(t, er)
	addrBz2, er := hex.DecodeString(accs[1].Address)
	require.NoError(t, er)
	msg := NewTestingSendMessage(t, addrBz, addrBz2, sendAmountString)
	candidates, err := ctx.GetMessageSendSignerCandidates(&msg)
	require.NoError(t, err)
<<<<<<< HEAD
	require.True(t, len(candidates) == 1, fmt.Sprintf("wrong number of candidates, expected %d, got %d", 1, len(candidates)))
	require.True(t, bytes.Equal(candidates[0], addrBz), fmt.Sprintf("unexpected signer candidate"))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.Equal(t, len(candidates), 1, "wrong number of candidates")
	require.Equal(t, candidates[0], accs[0].Address, "unexpected signer candidate")

	tests.CleanupTest(ctx)
>>>>>>> main
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
<<<<<<< HEAD
	require.True(t, amount == gotAmountString, fmt.Sprintf("unexpected amount, expected %s got %s", amount, gotAmountString))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.Equal(t, amount, gotAmountString, "unexpected amount")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_SetAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	require.NoError(t, ctx.SetAccountAmount(addr, amount), "set account amount")
<<<<<<< HEAD
	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.True(t, gotAmount.Cmp(amount) == 0, fmt.Sprintf("unexpected amounts: expected %v, got %v", amount, gotAmount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount, "unexpected amounts")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_SetAccountWithAmountString(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	amount := big.NewInt(100)
	amountString := types.BigIntToString(amount)
	require.NoError(t, ctx.SetAccountWithAmountString(addr, amountString), "set account amount string")
<<<<<<< HEAD
	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.True(t, gotAmount.Cmp(amount) == 0, fmt.Sprintf("unexpected amounts: expected %v, got %v", amount, gotAmount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======

	gotAmount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, amount, gotAmount, "unexpected amounts")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_SetPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]
<<<<<<< HEAD
	beforeAmount := pool.Amount
=======

	beforeAmount := pool.Account.Amount
>>>>>>> main
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	expectedAfterAmount := big.NewInt(100)
<<<<<<< HEAD
	require.NoError(t, ctx.SetPoolAmount(pool.Address, expectedAfterAmount), "set pool amount")
	amount, err := ctx.GetPoolAmount(pool.Address)
	require.NoError(t, err)
	require.True(t, beforeAmountBig.Cmp(amount) != 0, fmt.Sprintf("no amount change in pool"))
	require.True(t, expectedAfterAmount.Cmp(amount) == 0, fmt.Sprintf("unexpected pool amount; expected %v got %v", expectedAfterAmount, amount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.NoError(t, ctx.SetPoolAmount(pool.Name, expectedAfterAmount), "set pool amount")

	amount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)
	require.Equal(t, beforeAmountBig, defaultAmount, "no amount change in pool")
	require.Equal(t, expectedAfterAmount, amount, "unexpected pool amount")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_SubPoolAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pool := GetAllTestingPools(t, ctx)[0]

	beforeAmountBig := big.NewInt(1000000000000000)
<<<<<<< HEAD
	ctx.SetPoolAmount(pool.Address, beforeAmountBig)
	subAmountBig := big.NewInt(100)
	subAmount := types.BigIntToString(subAmountBig)
	require.NoError(t, ctx.SubPoolAmount(pool.Address, subAmount), "sub pool amount")
	amount, err := ctx.GetPoolAmount(pool.Address)
	require.NoError(t, err)
	require.True(t, beforeAmountBig.Cmp(amount) != 0, fmt.Sprintf("no amount change in pool"))
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.True(t, expected.Cmp(amount) == 0, fmt.Sprintf("unexpected pool amount; expected %v got %v", expected, amount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	ctx.SetPoolAmount(pool.Name, beforeAmountBig)

	subAmountBig := big.NewInt(100)
	subAmount := types.BigIntToString(subAmountBig)
	require.NoError(t, ctx.SubPoolAmount(pool.Name, subAmount), "sub pool amount")

	amount, err := ctx.GetPoolAmount(pool.Name)
	require.NoError(t, err)

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, amount, expected, "unexpected pool amount")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func TestUtilityContext_SubtractAccountAmount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	acc := GetAllTestingAccounts(t, ctx)[0]

	beforeAmount := acc.Amount
	beforeAmountBig, err := types.StringToBigInt(beforeAmount)
	require.NoError(t, err)

	subAmountBig := big.NewInt(100)
<<<<<<< HEAD
	addrBz, er := hex.DecodeString(acc.Address)
	require.NoError(t, er)
	require.NoError(t, ctx.SubtractAccountAmount(addrBz, subAmountBig), "sub account amount")
	amount, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)
	require.True(t, beforeAmountBig.Cmp(amount) != 0, fmt.Sprintf("no amount change in pool"))
	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.True(t, expected.Cmp(amount) == 0, fmt.Sprintf("unexpected acc amount; expected %v got %v", expected, amount))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
=======
	require.NoError(t, ctx.SubtractAccountAmount(acc.Address, subAmountBig), "sub account amount")

	amount, err := ctx.GetAccountAmount(acc.Address)
	require.NoError(t, err)
	require.Equal(t, beforeAmountBig, defaultAmount, "no amount change in pool")

	expected := beforeAmountBig.Sub(beforeAmountBig, subAmountBig)
	require.Equal(t, expected, amount, "unexpected acc amount")

	tests.CleanupTest(ctx)
>>>>>>> main
}

func GetAllTestingAccounts(t *testing.T, ctx utility.UtilityContext) []*genesis.Account {
	accs, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllAccounts(0)
	sort.Slice(accs, func(i, j int) bool {
<<<<<<< HEAD
		return accs[i].Address < accs[j].Address
=======
		return hex.EncodeToString(accs[i].Address) < hex.EncodeToString(accs[j].Address)
>>>>>>> main
	})
	require.NoError(t, err)
	return accs
}

<<<<<<< HEAD
func GetAllTestingPools(t *testing.T, ctx utility.UtilityContext) []*genesis.Account {
	accs, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllPools(0)
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].Address < accs[j].Address
=======
func GetAllTestingPools(t *testing.T, ctx utility.UtilityContext) []*genesis.Pool {
	accs, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllPools(0)
	sort.Slice(accs, func(i, j int) bool {
		return accs[i].Name < accs[j].Name
>>>>>>> main
	})
	require.NoError(t, err)
	return accs
}
