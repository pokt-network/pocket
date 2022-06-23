package test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
)

func FuzzAccountAmount(f *testing.F) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	ops := []string{"Add", "Sub", "Set", "NextHeight"}
	acc := NewTestAccount(nil)
	db.SetAccountAmount(acc.Address, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())
	var delta *big.Int
	numOptions := len(ops)
	numOperations := 20
	for i := 0; i < numOperations; i++ {
		f.Add(ops[rand.Intn(numOptions)])
	}
	f.Fuzz(func(t *testing.T, op string) {
		delta = big.NewInt(int64(rand.Intn(1000)))
		switch op {
		case "Add":
			originalAmountBig, err := db.GetAccountAmount(acc.Address, db.Height)
			require.NoError(t, err)
			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)
			err = db.AddAccountAmount(acc.Address, types.BigIntToString(delta))
			require.NoError(t, err)
			expectedAmount.Add(originalAmount, delta)
		case "Sub":
			originalAmountBig, err := db.GetAccountAmount(acc.Address, db.Height)
			require.NoError(t, err)
			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)
			err = db.SubtractAccountAmount(acc.Address, types.BigIntToString(delta))
			require.NoError(t, err)
			expectedAmount.Sub(originalAmount, delta)
		case "Set":
			err := db.SetAccountAmount(acc.Address, types.BigIntToString(delta))
			require.NoError(t, err)
			expectedAmount = delta
		case "NextHeight":
			db.Height++
		}
		currentAmount, err := db.GetAccountAmount(acc.Address, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

func TestSetAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)
	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)
	am, err := db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, am, "unexpected amount")
	db.SetAccountAmount(acc.Address, StakeToUpdate)
	require.NoError(t, err)
	am, err = db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, am, "unexpected amount after second set")
}

func TestAddAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)
	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)
	amountToAddBig := big.NewInt(100)
	err = db.AddAccountAmount(acc.Address, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)
	am, err := db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)
	resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after add")
}

func TestSubAccountAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	acc := NewTestAccount(t)
	err := db.SetAccountAmount(acc.Address, DefaultStake)
	require.NoError(t, err)
	amountToSubBig := big.NewInt(100)
	db.SubtractAccountAmount(acc.Address, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)
	am, err := db.GetAccountAmount(acc.Address, db.Height)
	require.NoError(t, err)
	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after sub")
}

func FuzzPoolAmount(f *testing.F) {
	// Setup
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	ops := []string{"Add", "Sub", "Set", "NextHeight"}
	acc := NewTestPool(nil)
	db.SetPoolAmount(acc.Name, DefaultAccountAmount)
	expectedAmount := big.NewInt(DefaultAccountBig.Int64())
	var delta *big.Int
	numOptions := len(ops)
	numOperations := 20
	for i := 0; i < numOperations; i++ {
		f.Add(ops[rand.Intn(numOptions)])
	}
	f.Fuzz(func(t *testing.T, op string) {
		delta = big.NewInt(int64(rand.Intn(1000)))
		switch op {
		case "Add":
			originalAmountBig, err := db.GetPoolAmount(acc.Name, db.Height)
			require.NoError(t, err)
			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)
			err = db.AddPoolAmount(acc.Name, types.BigIntToString(delta))
			require.NoError(t, err)
			expectedAmount.Add(originalAmount, delta)
		case "Sub":
			originalAmountBig, err := db.GetPoolAmount(acc.Name, db.Height)
			require.NoError(t, err)
			originalAmount, err := types.StringToBigInt(originalAmountBig)
			require.NoError(t, err)
			err = db.SubtractPoolAmount(acc.Name, types.BigIntToString(delta))
			require.NoError(t, err)
			expectedAmount.Sub(originalAmount, delta)
		case "Set":
			err := db.SetPoolAmount(acc.Name, types.BigIntToString(delta))
			require.NoError(t, err)
			expectedAmount = delta
		case "NextHeight":
			db.Height++
		}
		currentAmount, err := db.GetPoolAmount(acc.Name, db.Height)
		require.NoError(t, err)
		require.Equal(t, types.BigIntToString(expectedAmount), currentAmount, fmt.Sprintf("unexpected amount after %s", op))
	})
}

func TestSetPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)
	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)
	am, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, am, "unexpected amount")
	err = db.SetPoolAmount(pool.Name, StakeToUpdate)
	require.NoError(t, err)
	am, err = db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	require.Equal(t, StakeToUpdate, am, "unexpected amount after second set")
}

func TestAddPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)
	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)
	amountToAddBig := big.NewInt(100)
	err = db.AddPoolAmount(pool.Name, types.BigIntToString(amountToAddBig))
	require.NoError(t, err)
	am, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	resultBig := (&big.Int{}).Add(DefaultStakeBig, amountToAddBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after add")
}

func TestSubPoolAmount(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	pool := NewTestPool(t)
	err := db.SetPoolAmount(pool.Name, DefaultStake)
	require.NoError(t, err)
	amountToSubBig := big.NewInt(100)
	err = db.SubtractPoolAmount(pool.Name, types.BigIntToString(amountToSubBig))
	require.NoError(t, err)
	am, err := db.GetPoolAmount(pool.Name, db.Height)
	require.NoError(t, err)
	resultBig := (&big.Int{}).Sub(DefaultStakeBig, amountToSubBig)
	expectedResult := types.BigIntToString(resultBig)
	require.Equal(t, expectedResult, am, "unexpected amount after sub")
}

// --- Helpers ---

func NewTestAccount(t *testing.T) typesGenesis.Account {
	addr, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}
	return typesGenesis.Account{
		Address: addr,
		Amount:  DefaultAccountAmount,
	}
}

func NewTestPool(t *testing.T) typesGenesis.Pool {
	_, err := crypto.GenerateAddress()
	if t != nil {
		require.NoError(t, err)
	}
	return typesGenesis.Pool{
		Name: DefaultPoolName,
		Account: &typesGenesis.Account{
			Amount: DefaultAccountAmount,
		},
	}
}
