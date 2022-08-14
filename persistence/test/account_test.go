package test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzAccountAmount(f *testing.F) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}
	operations := []string{
		"AddAmount",
		"SubAmount",
		"SetAmount",

		"IncrementHeight",
	}
	numOperationTypes := len(operations)

	account := newTestAccount(nil)
	db.SetAccountAmount(account.Address, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())

	numDbOperations := 20
	for i := 0; i < numDbOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)])
	}

	f.Fuzz(func(t *testing.T, op string) {
		delta := big.NewInt(int64(rand.Intn(1000)))
		deltaString := types.BigIntToString(delta)

		switch op {
		case "AddAmount":
			originalAmountBig, err := db.GetAccountAmount(account.Address, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.AddAccountAmount(account.Address, deltaString)
			require.NoError(t, err)

			expectedAmount.Add(originalAmount, delta)
		case "SubAmount":
			originalAmountBig, err := db.GetAccountAmount(account.Address, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.SubtractAccountAmount(account.Address, deltaString)
			require.NoError(t, err)

			expectedAmount.Sub(originalAmount, delta)
		case "SetAmount":
			err := db.SetAccountAmount(account.Address, deltaString)
			require.NoError(t, err)

			expectedAmount = delta
		case "IncrementHeight":
			db.Height++
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}

		currentAmount, err := db.GetAccountAmount(account.Address, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

func TestSetAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	account := newTestAccount(t)

	err := db.SetAccountAmount(account.Address, DefaultStake)
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(account.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, accountAmount, "unexpected amount")

	err = db.SetAccountAmount(account.Address, StakeToUpdate)
	require.NoError(t, err)

	accountAmount, err = db.GetAccountAmount(account.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, accountAmount, "unexpected amount after second set")
}

func TestAddAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	account := newTestAccount(t)

	err := db.SetAccountAmount(account.Address, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.AddAccountAmount(account.Address, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(account.Address, db.Height)
	require.NoError(t, err)

	accountAmountBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedAccountAmount := types.BigIntToString(accountAmountBig)

	require.Equal(t, expectedAccountAmount, accountAmount, "unexpected amount after add")
}

func TestSubAccountAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	account := newTestAccount(t)

	err := db.SetAccountAmount(account.Address, DefaultStake)
	require.NoError(t, err)

	amountToSubBig := big.NewInt(100)
	err = db.SubtractAccountAmount(account.Address, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)

	accountAmount, err := db.GetAccountAmount(account.Address, db.Height)
	require.NoError(t, err)

	accountAmountBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedAccountAmount := types.BigIntToString(accountAmountBig)
	require.Equal(t, expectedAccountAmount, accountAmount, "unexpected amount after sub")
}

func FuzzPoolAmount(f *testing.F) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *testPostgresDB,
	}

	operations := []string{
		"AddAmount",
		"SubAmount",
		"SetAmount",

		"IncrementHeight",
	}
	numOperationTypes := len(operations)

	pool := newTestPool(nil)
	db.SetPoolAmount(pool.Name, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())

	numDbOperations := 20
	for i := 0; i < numDbOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)])
	}

	f.Fuzz(func(t *testing.T, op string) {
		delta := big.NewInt(int64(rand.Intn(1000)))
		deltaString := types.BigIntToString(delta)

		switch op {
		case "AddAmount":
			originalAmountBig, err := db.GetPoolAmount(pool.Name, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.AddPoolAmount(pool.Name, deltaString)
			require.NoError(t, err)

			expectedAmount.Add(originalAmount, delta)
		case "SubAmount":
			originalAmountBig, err := db.GetPoolAmount(pool.Name, db.Height)
			require.NoError(t, err)

			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)

			err = db.SubtractPoolAmount(pool.Name, deltaString)
			require.NoError(t, err)

			expectedAmount.Sub(originalAmount, delta)
		case "SetAmount":
			err := db.SetPoolAmount(pool.Name, deltaString)
			require.NoError(t, err)

			expectedAmount = delta
		case "IncrementHeight":
			db.Height++
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}

		currentAmount, err := db.GetPoolAmount(pool.Name, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

func TestSetPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	pool := newTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	poolAmount, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, poolAmount, "unexpected amount")

	err = db.SetPoolAmount(pool.Name, StakeToUpdate)
	require.NoError(t, err)

	poolAmount, err = db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, poolAmount, "unexpected amount after second set")
}

func TestAddPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	pool := newTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	amountToAddBig := big.NewInt(100)
	err = db.AddPoolAmount(pool.Name, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)

	poolAmount, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)

	poolAmountBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedPoolAmount := types.BigIntToString(poolAmountBig)

	require.Equal(t, expectedPoolAmount, poolAmount, "unexpected amount after add")
}

func TestSubPoolAmount(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	pool := newTestPool(t)

	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)

	amountToSubBig := big.NewInt(100)
	err = db.SubtractPoolAmount(pool.Name, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)

	poolAmount, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)

	poolAmountBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedPoolAmount := types.BigIntToString(poolAmountBig)
	require.Equal(t, expectedPoolAmount, poolAmount, "unexpected amount after sub")
}

func TestGetAllAccounts(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	updateAccount := func(db *persistence.PostgresContext, acc *genesis.Account) error {
		return db.AddAccountAmount(acc.Address, "10")
	}

	getAllActorsTest(t, db, db.GetAllAccounts, createAndInsertNewAccount, updateAccount, 9)
}

func TestGetAllPools(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	updatePool := func(db *persistence.PostgresContext, pool *genesis.Pool) error {
		return db.AddPoolAmount(pool.Name, "10")
	}

	getAllActorsTest(t, db, db.GetAllPools, createAndInsertNewPool, updatePool, 6)
}

// --- Helpers ---

func createAndInsertNewAccount(db *persistence.PostgresContext) (*genesis.Account, error) {
	account := newTestAccount(nil)
	return &account, db.SetAccountAmount(account.Address, DefaultAccountAmount)
}

func newTestAccount(t *testing.T) typesGenesis.Account {
	addr, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}
	return typesGenesis.Account{
		Address: addr,
		Amount:  DefaultAccountAmount,
	}
}

func createAndInsertNewPool(db *persistence.PostgresContext) (*genesis.Pool, error) {
	pool := newTestPool(nil)
	return &pool, db.SetPoolAmount(pool.Name, DefaultAccountAmount)
}

func newTestPool(t *testing.T) typesGenesis.Pool {
	_, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}
	return typesGenesis.Pool{
		Name: fmt.Sprintf("%s_%d", DefaultPoolName, rand.Int()),
		Account: &typesGenesis.Account{
			Amount: DefaultAccountAmount,
		},
	}
}
